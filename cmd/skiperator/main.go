package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/api/v1beta1"
	"github.com/kartverket/skiperator/internal/config"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/imagepullsecret"
	"github.com/kartverket/skiperator/pkg/resourcegenerator/networkpolicy/defaultdeny"
	"github.com/kartverket/skiperator/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kartverket/skiperator/internal/controllers"
	"github.com/kartverket/skiperator/internal/controllers/common"
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

	realzap "go.uber.org/zap"
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
	encCfg := realzap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// Set a temporary logger for the config loading, the real logger is initialized later with values from config
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Encoder:     zapcore.NewJSONEncoder(encCfg),
		Development: true,
		Level:       zapcore.InfoLevel,
		DestWriter:  os.Stdout,
	})))

	kubeconfig := ctrl.GetConfigOrDie()

	configClient, err := client.New(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "could not create config check client")
		os.Exit(1)
	}

	// If loading configuration takes over 15 seconds, something is seriously wrong.
	// We should let skiperator crash and let a new process attempt to load the configuration again.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// load global config. exit on error
	err = config.LoadConfig(ctx, configClient)
	if err != nil {
		setupLog.Error(err, "could not load global config")
		os.Exit(1)
	} else {
		setupLog.Info("Successfully loaded global config")
	}

	activeConfig := config.GetActiveConfig()

	setupLog.Info(fmt.Sprintf("Running skiperator %s (commit %s), with %d concurrent reconciles", Version, Commit, activeConfig.ConcurrentReconciles))

	parsedLogLevel, _ := zapcore.ParseLevel(activeConfig.LogLevel)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Encoder:     zapcore.NewJSONEncoder(encCfg),
		Development: !activeConfig.IsDeployment,
		Level:       parsedLogLevel,
		DestWriter:  os.Stdout,
	})))

	if !activeConfig.IsDeployment && !strings.Contains(kubeconfig.Host, "https://127.0.0.1") {
		setupLog.Info("Tried to start skiperator with non-local kubecontext. Exiting to prevent havoc.")
		os.Exit(1)
	} else {
		setupLog.Info(fmt.Sprintf("Starting skiperator using kube-apiserver at %s", kubeconfig.Host))
	}

	detectK8sVersion(kubeconfig)

	pprofBindAddr := ""

	if activeConfig.EnableProfiling {
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
		LeaderElection:          activeConfig.LeaderElection,
		LeaderElectionNamespace: activeConfig.LeaderElectionNamespace,
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

	err = (&controllers.ApplicationReconciler{
		ReconcilerBase:   common.NewFromManager(mgr, mgr.GetEventRecorderFor("application-controller")),
		SkiperatorConfig: activeConfig,
	}).SetupWithManager(mgr, activeConfig.ConcurrentReconciles)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	var skipClusterList *config.SKIPClusterList
	if activeConfig.ClusterCIDRExclusionEnabled {
		err = config.ValidateSKIPClusterList(&activeConfig.ClusterCIDRMap)
		if err != nil {
			setupLog.Error(err, "could not load SKIP cluster config")
			os.Exit(1)
		} else {
			skipClusterList = &activeConfig.ClusterCIDRMap
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
		ReconcilerBase:   common.NewFromManager(mgr, mgr.GetEventRecorderFor("skipjob-controller")),
		SkiperatorConfig: activeConfig,
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
	var regSecrets []imagepullsecret.RegistryCredentialSecret
	for _, registry := range activeConfig.RegistrySecretRefs {
		secret, regErr := util.GetSecret(configClient, ctx, types.NamespacedName{
			Namespace: "skiperator-system",
			Name:      registry.SecretName,
		})
		if regErr != nil {
			setupLog.Error(err, "unable to fetch registry credential secret configuration", "controller", "Namespace")
			os.Exit(1)
		} else {
			regcredSecret := imagepullsecret.RegistryCredentialSecret{
				Registry:  registry.Registry,
				Secret:    secret,
				SecretKey: registry.SecretKey,
			}
			regSecrets = append(regSecrets, regcredSecret)
		}
	}
	ps, err := imagepullsecret.NewImagePullSecret(regSecrets...)
	if err != nil {
		setupLog.Error(err, "unable to create image pull secret configuration", "controller", "Namespace")
		os.Exit(1)
	}
	setupLog.Info("initialized image pull secret", "controller", "Namespace", "registry-count", len(activeConfig.RegistrySecretRefs))

	dd, err := defaultdeny.NewDefaultDenyNetworkPolicy(skipClusterList, activeConfig.ClusterCIDRExclusionEnabled)
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

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

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
