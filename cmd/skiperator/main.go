package main

import (
	"flag"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	"os"

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
	"github.com/kartverket/skiperator/controllers"
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
}

func main() {
	leaderElection := flag.Bool("l", false, "enable leader election")
	leaderElectionNamespace := flag.String("ln", "", "leader election namespace")
	imagePullToken := flag.String("t", "", "image pull token")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: true})))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		HealthProbeBindAddress:  ":8081",
		LeaderElection:          *leaderElection,
		LeaderElectionNamespace: *leaderElectionNamespace,
		LeaderElectionID:        "skiperator",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = (&controllers.ServiceAccountReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServiceAccount")
		os.Exit(1)
	}

	err = (&controllers.ImagePullSecretReconciler{Registry: "ghcr.io", Token: *imagePullToken}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ImagePullSecret")
		os.Exit(1)
	}

	err = (&controllers.DeploymentReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}

	err = (&controllers.HorizontalPodAutoscalerReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HorizontalPodAutoscaler")
		os.Exit(1)
	}

	err = (&controllers.DefaultDenyNetworkPolicyReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DefaultDenyNetworkPolicy")
		os.Exit(1)
	}

	err = (&controllers.NetworkPolicyReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkPolicy")
		os.Exit(1)
	}

	err = (&controllers.ServiceReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Service")
		os.Exit(1)
	}

	err = (&controllers.SidecarReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Sidecar")
		os.Exit(1)
	}

	err = (&controllers.PeerAuthenticationReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PeerAuthentication")
		os.Exit(1)
	}

	err = (&controllers.CertificateReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Certificate")
		os.Exit(1)
	}

	err = (&controllers.IngressGatewayReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IngressGateway")
		os.Exit(1)
	}

	err = (&controllers.IngressVirtualServiceReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IngressVirtualService")
		os.Exit(1)
	}

	err = (&controllers.EgressServiceEntryReconciler{}).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "EgressServiceEntry")
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
