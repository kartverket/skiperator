package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kartverket/skiperator/internal/controllers"
	"github.com/kartverket/skiperator/internal/controllers/common"
	"github.com/kartverket/skiperator/pkg/flags"
	"github.com/kartverket/skiperator/pkg/k8sfeatures"
	"github.com/kartverket/skiperator/pkg/log"
	"github.com/kartverket/skiperator/pkg/metrics/usage"
	"github.com/kartverket/skiperator/pkg/resourceschemas"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"go.uber.org/zap/zapcore"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
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
	resourceschemas.AddSchemas(scheme)
}

func main() {
	leaderElection := flag.Bool("l", false, "enable leader election")
	leaderElectionNamespace := flag.String("ln", "", "leader election namespace")
	imagePullToken := flag.String("t", "", "image pull token")
	isDeployment := flag.Bool("d", false, "is deployed to a real cluster")
	logLevel := flag.String("e", "debug", "Error level used for logs. Default debug. Possible values: debug, info, warn, error, dpanic, panic.")
	flag.Parse()

	parsedLogLevel, _ := zapcore.ParseLevel(*logLevel)

	//TODO use zap directly so we get more loglevels
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: !*isDeployment,
		Level:       parsedLogLevel,
		DestWriter:  os.Stdout,
	})))

	setupLog.Info(fmt.Sprintf("Running skiperator %s (commit %s)", Version, Commit))

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

	mgr, err := ctrl.NewManager(kubeconfig, ctrl.Options{
		Scheme:                  scheme,
		HealthProbeBindAddress:  ":8081",
		LeaderElection:          *leaderElection,
		LeaderElectionNamespace: *leaderElectionNamespace,
		Metrics:                 metricsserver.Options{BindAddress: ":8181"},
		LeaderElectionID:        "skiperator",
		PprofBindAddress:        pprofBindAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := usage.NewUsageMetrics(kubeconfig, log.NewLogger().WithName("usage-metrics")); err != nil {
		setupLog.Error(err, "unable to configure usage metrics")
		os.Exit(1)
	}

	err = (&controllers.ApplicationReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("application-controller"), resourceschemas.GetApplicationSchemas(mgr.GetScheme())),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	err = (&controllers.SKIPJobReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("skipjob-controller"), resourceschemas.GetJobSchemas(mgr.GetScheme())),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SKIPJob")
		os.Exit(1)
	}

	err = (&controllers.RoutingReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("routing-controller"), resourceschemas.GetRoutingSchemas(mgr.GetScheme())),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Routing")
		os.Exit(1)
	}

	err = (&controllers.NamespaceReconciler{
		ReconcilerBase: common.NewFromManager(mgr, mgr.GetEventRecorderFor("namespace-controller"), resourceschemas.GetNamespaceSchemas(mgr.GetScheme())),
		Registry:       "ghcr.io",
		Token:          *imagePullToken,
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
