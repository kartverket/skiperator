package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/internal/controllers"
	"github.com/kartverket/skiperator/internal/controllers/common"
	webhookv1beta1 "github.com/kartverket/skiperator/internal/webhook"
	"github.com/kartverket/skiperator/pkg/envconfig"
	"github.com/kartverket/skiperator/pkg/flags"
	"github.com/kartverket/skiperator/pkg/k8sfeatures"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/metrics/usage"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/imagepullsecret"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/defaultdeny"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"go.uber.org/zap/zapcore"
	istioclientv1 "istio.io/client-go/pkg/apis/networking/v1"
	istioclientv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istioclienttelemetryv1 "istio.io/client-go/pkg/apis/telemetry/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/caarlos0/env/v11"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;create;update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	Version = "dev"
	Commit  = "N/A"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	// Register Istio types
	utilruntime.Must(istioclientv1.AddToScheme(scheme))
	utilruntime.Must(istioclientv1beta1.AddToScheme(scheme))
	utilruntime.Must(istioclienttelemetryv1.AddToScheme(scheme))
	resourceschemas.AddSchemas(scheme)
}

func main() {
	leaderElection := flag.Bool("l", false, "enable leader election")
	leaderElectionNamespace := flag.String("ln", "", "leader election namespace")
	isDeployment := flag.Bool("d", false, "is deployed to a real cluster")
	logLevel := flag.String("e", "debug", "Error level used for logs. Default debug. Possible values: debug, info, warn, error, dpanic, panic.")
	concurrentReconciles := flag.Int("c", 1, "number of concurrent reconciles for application controller")
	webhookCertDir := flag.String("webhook-cert-dir", "", "Directory containing webhook TLS certificate and key. If empty, webhook will use self-signed certificates.")
	webhookCertName := flag.String("webhook-cert-name", "tls.crt", "Name of the webhook TLS certificate file")
	webhookKeyName := flag.String("webhook-key-name", "tls.key", "Name of the webhook TLS key file")
	webhookHost := flag.String("webhook-host", "", "Host to bind webhook server to. Use 0.0.0.0 to bind to all interfaces for local kind development.")
	webhookPort := flag.Int("webhook-port", 9443, "Port for webhook server to listen on")
	flag.Parse()

	// Providing multiple image pull tokens as flags are painful, so instead we parse them as env variables

	parsedLogLevel, _ := zapcore.ParseLevel(*logLevel)

	//TODO use zap directly so we get more loglevels
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: !*isDeployment,
		Level:       parsedLogLevel,
		DestWriter:  os.Stdout,
	})))

	var cfg envconfig.Vars
	if parseErr := env.Parse(&cfg); parseErr != nil {
		setupLog.Error(parseErr, "Failed to parse config")
		os.Exit(1)
	}

	setupLog.Info(fmt.Sprintf("Running skiperator %s (commit %s), with %d concurrent reconciles", Version, Commit, *concurrentReconciles))

	kubeconfig := ctrl.GetConfigOrDie()

	if !*isDeployment && !strings.Contains(kubeconfig.Host, "https://127.0.0.1") {
		setupLog.Info("Tried to start skiperator with non-local kubecontext. Exiting to prevent havoc.")
		os.Exit(1)
	} else {
		setupLog.Info(fmt.Sprintf("Starting skiperator using kube-apiserver at %s", kubeconfig.Host))
	}

	detectK8sVersion(kubeconfig)

	pprofBindAddr := ""
	if flags.FeatureFlags.EnableProfiling {
		pprofBindAddr = ":8281"
	}
	var metricsAddr string
	var metricsCertPath, metricsCertName, metricsCertKey string
	// Create watchers for metrics and webhooks certificates
	var metricsCertWatcher, webhookCertWatcher *certwatcher.CertWatcher
	var secureMetrics bool
	var tlsOpts []func(*tls.Config)
	// Initial webhook TLS options
	webhookTLSOpts := tlsOpts

	// Initialize webhook certificate watcher if webhook cert directory is provided
	if len(*webhookCertDir) > 0 {
		certPath := filepath.Join(*webhookCertDir, *webhookCertName)
		keyPath := filepath.Join(*webhookCertDir, *webhookKeyName)

		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-dir", *webhookCertDir, "webhook-cert-name", *webhookCertName, "webhook-key-name", *webhookKeyName)

		var err error
		webhookCertWatcher, err = certwatcher.New(certPath, keyPath)
		if err != nil {
			setupLog.Error(err, "Failed to initialize webhook certificate watcher")
			os.Exit(1)
		}

		webhookTLSOpts = append(webhookTLSOpts, func(config *tls.Config) {
			config.GetCertificate = webhookCertWatcher.GetCertificate
		})
	} else {
		setupLog.Info("No webhook certificate directory provided, webhook will use self-signed certificates")
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: webhookTLSOpts,
		Host:    *webhookHost,
		Port:    *webhookPort,
	})

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/metrics/filters#WithAuthenticationAndAuthorization
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// If the certificate is not specified, controller-runtime will automatically
	// generate self-signed certificates for the metrics server. While convenient for development and testing,

	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		var err error
		metricsCertWatcher, err = certwatcher.New(
			filepath.Join(metricsCertPath, metricsCertName),
			filepath.Join(metricsCertPath, metricsCertKey),
		)
		if err != nil {
			setupLog.Error(err, "to initialize metrics certificate watcher", "error", err)
			os.Exit(1)
		}

		metricsServerOptions.TLSOpts = append(metricsServerOptions.TLSOpts, func(config *tls.Config) {
			config.GetCertificate = metricsCertWatcher.GetCertificate
		})
	}
	// Create new manager
	mgr, err := ctrl.NewManager(kubeconfig, ctrl.Options{
		Scheme:                  scheme,
		HealthProbeBindAddress:  ":8081",
		LeaderElection:          *leaderElection,
		LeaderElectionNamespace: *leaderElectionNamespace,
		Metrics:                 metricsserver.Options{BindAddress: ":8181"},
		LeaderElectionID:        "skiperator",
		PprofBindAddress:        pprofBindAddr,
		WebhookServer:           webhookServer,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	// Setup leader-specific tasks as a goroutine that runs after manager starts
	go func() {
		<-mgr.Elected() // Wait until this instance is elected as leader
		setupLog.Info("I am the captain now – configuring usage metrics")
		if err := usage.NewUsageMetrics(kubeconfig, log.NewLogger().WithName("usage-metrics")); err != nil {
			setupLog.Error(err, "unable to configure usage metrics")
		}
	}()

	// Setup webhooks first
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err := webhookv1beta1.SetupSkipJobWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "SvartSkjaif")
			os.Exit(1)
		}
	}

	// Add certificate watchers
	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			setupLog.Error(err, "unable to add metrics certificate watcher to manager")
			os.Exit(1)
		}
	}

	if webhookCertWatcher != nil {
		setupLog.Info("Adding webhook certificate watcher to manager")
		if err := mgr.Add(webhookCertWatcher); err != nil {
			setupLog.Error(err, "unable to add webhook certificate watcher to manager")
			os.Exit(1)
		}
	}

	// Setup health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	configCheckClient, err := client.New(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "could not create config check client")
		os.Exit(1)
	}

	// If loading configuration takes over 15 seconds, something is seriously wrong.
	// We should let skiperator crash and let a new process attempt to load the configuration again.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// load global config. exit on error
	err = config.LoadConfig(ctx, configCheckClient)
	if err != nil {
		setupLog.Error(err, "could not load global config")
		os.Exit(1)
	} else {
		setupLog.Info("Successfully loaded global config", "config", config.GetActiveConfig())
	}

	// We need to get this configmap before initializing the manager, therefore we need a separate client for thi
	// If the configmap is not present or otherwise misconfigured, we should not start Skiperator
	// as this configmap contains the CIDRs for cluster nodes in order to prevent egress traffic
	// directly from namespaces as per SKIP-1704

	var skipClusterList *config.SKIPClusterList
	if cfg.ClusterCIDRExclusionEnabled {
		skipClusterList, err = config.LoadSKIPClusterConfigFromConfigMap(configCheckClient)
		if err != nil {
			setupLog.Error(err, "could not load SKIP cluster config")
			os.Exit(1)
		}
	} else {
		skipClusterList = &config.SKIPClusterList{
			Clusters: []*config.SKIPCluster{},
		}
	}

	// Setup all controllers
	err = (&controllers.ApplicationReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("application-controller")),
	}).SetupWithManager(mgr, concurrentReconciles)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	err = (&controllers.SKIPJobReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("skipjob-controller")),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SKIPJob")
		os.Exit(1)
	}

	err = (&controllers.RoutingReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("routing-controller")),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Routing")
		os.Exit(1)
	}

	ps, err := imagepullsecret.NewImagePullSecret(cfg.RegistryCredentials...)
	if err != nil {
		setupLog.Error(err, "unable to create image pull secret configuration", "controller", "Namespace")
		os.Exit(1)
	}
	setupLog.Info("initialized image pull secret", "controller", "Namespace", "registry-count", len(cfg.RegistryCredentials))

	dd, err := defaultdeny.NewDefaultDenyNetworkPolicy(skipClusterList)
	if err != nil {
		setupLog.Error(err, "unable to create default deny network policy configuration", "controller", "Namespace")
		os.Exit(1)
	}

	err = (&controllers.NamespaceReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("namespace-controller")),
		PullSecret:     ps,
		DefaultDeny:    dd,
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}

	if flags.FeatureFlags == nil {
		panic("something is wrong with the go runtime, panicing")
	}
	setupLog.Info("initializing skiperator with feature flags", "features", flags.FeatureFlags)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func detectK8sVersion(kubeconfig *rest.Config) {
	disco, err := discovery.NewDiscoveryClientForConfig(kubeconfig)
	if err != nil {
		setupLog.Error(err, "could not get discovery client")
		os.Exit(1)
	}
	ver, err := disco.ServerVersion()
	if err != nil {
		setupLog.Error(err, "could not get server version")
		os.Exit(1)
	}
	k8sfeatures.NewVersionInfo(ver)
	setupLog.Info("detected server version", "version", ver)
}
