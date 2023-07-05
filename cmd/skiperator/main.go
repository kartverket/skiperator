package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	pov1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.uber.org/zap/zapcore"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	policyv1 "k8s.io/api/policy/v1"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	applicationcontroller "github.com/kartverket/skiperator/controllers/application"
	namespacecontroller "github.com/kartverket/skiperator/controllers/namespace"
	"github.com/kartverket/skiperator/pkg/util"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
)

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;create;update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(skiperatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(autoscalingv2.AddToScheme(scheme))
	utilruntime.Must(securityv1beta1.AddToScheme(scheme))
	utilruntime.Must(networkingv1beta1.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1.AddToScheme(scheme))
	utilruntime.Must(policyv1.AddToScheme(scheme))
	utilruntime.Must(pov1.AddToScheme(scheme))
	utilruntime.Must(nais_io_v1.AddToScheme(scheme))
}

func main() {
	leaderElection := flag.Bool("l", false, "enable leader election")
	leaderElectionNamespace := flag.String("ln", "", "leader election namespace")
	imagePullToken := flag.String("t", "", "image pull token")
	isDeployment := flag.Bool("d", false, "is deployed to a real cluster")
	logLevel := flag.String("e", "debug", "Error level used for logs. Default debug. Possible values: debug, info, warn, error, dpanic, panic.")
	flag.Parse()

	parsedLogLevel, _ := zapcore.ParseLevel(*logLevel)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: !*isDeployment,
		Level:       parsedLogLevel,
	})))

	kubeconfig := ctrl.GetConfigOrDie()

	if !*isDeployment && !strings.Contains(kubeconfig.Host, "https://127.0.0.1") {
		setupLog.Info("Tried to start skiperator with non-local kubecontext. Exiting to prevent havoc.")
		os.Exit(1)
	} else {
		setupLog.Info(fmt.Sprintf("Starting skiperator using kube-apiserver at %s", kubeconfig.Host))
	}

	mgr, err := ctrl.NewManager(kubeconfig, ctrl.Options{
		Scheme:                  scheme,
		HealthProbeBindAddress:  ":8081",
		LeaderElection:          *leaderElection,
		LeaderElectionNamespace: *leaderElectionNamespace,
		MetricsBindAddress:      ":8181",
		LeaderElectionID:        "skiperator",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = (&applicationcontroller.ApplicationReconciler{
		ReconcilerBase: util.NewFromManager(mgr, mgr.GetEventRecorderFor("application-controller")),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	err = (&namespacecontroller.NamespaceReconciler{
		ReconcilerBase: util.NewFromManager(mgr, mgr.GetEventRecorderFor("namespace-controller")),
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

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
