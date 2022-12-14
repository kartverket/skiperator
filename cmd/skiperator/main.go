package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	skiperatorv1alpha1 "github.com/kartverket/skiperator/api/v1alpha1"
	"github.com/kartverket/skiperator/controllers"
	skipmetrics "github.com/kartverket/skiperator/pkg/metrics"
	"github.com/kartverket/skiperator/pkg/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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
	utilruntime.Must(autoscalingv2beta2.AddToScheme(scheme))
	utilruntime.Must(securityv1beta1.AddToScheme(scheme))
	utilruntime.Must(networkingv1beta1.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
}

func main() {
	leaderElection := flag.Bool("l", false, "enable leader election")
	leaderElectionNamespace := flag.String("ln", "", "leader election namespace")
	imagePullToken := flag.String("t", "", "image pull token")
	isDeployment := flag.Bool("d", false, "is deployed to a real cluster")
	probeAddr := flag.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to")
	metricsAddr := flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: true})))

	kubeconfig := ctrl.GetConfigOrDie()

	if !*isDeployment && !strings.Contains(kubeconfig.Host, "https://127.0.0.1") {
		setupLog.Info("Tried to start skiperator with non-local kubecontext. Exiting to prevent havoc.")
		os.Exit(1)
	} else {
		setupLog.Info(fmt.Sprintf("Starting skiperator using kube-apiserver at %s", kubeconfig.Host))
	}

	skipmetrics.Register(metrics.Registry)
	mgr, err := ctrl.NewManager(kubeconfig, ctrl.Options{
		Scheme:                  scheme,
		HealthProbeBindAddress:  *probeAddr,
		MetricsBindAddress:      *metricsAddr,
		LeaderElection:          *leaderElection,
		LeaderElectionNamespace: *leaderElectionNamespace,
		LeaderElectionID:        "skiperator",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = (&controllers.ApplicationReconciler{
		ReconcilerBase: util.NewFromManager(mgr, mgr.GetEventRecorderFor("application-controller")),
	}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Application")
		os.Exit(1)
	}

	// These controllers may be added into application, but are not added due to currently being owned by namespaces, and not the application
	err = (&controllers.ImagePullSecretReconciler{Registry: "ghcr.io", Token: *imagePullToken}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ImagePullSecret")
		os.Exit(1)
	}

	err = (&controllers.DefaultDenyNetworkPolicyReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DefaultDenyNetworkPolicy")
		os.Exit(1)
	}

	err = (&controllers.SidecarReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Sidecar")
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
