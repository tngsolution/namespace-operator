package main

import (
	"os"

	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"
	"github.com/tngs/namespace-operator/controllers"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	networkingv1 "k8s.io/api/networking/v1" // 🔥 IMPORTANT

	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme))     // 🔥 REQUIRED
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme)) // Tenant & TenantProfile
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,

		Metrics: metricsserver.Options{
			BindAddress: ":8080",
		},

		HealthProbeBindAddress: ":8081",

		LeaderElection:   true,
		LeaderElectionID: "namespace-operator.platform.example.com",
	})

	if err != nil {
		os.Exit(1)
	}

	// ---------------------------------------------------------------------
	// Tenant controller
	// ---------------------------------------------------------------------
	if err = (&controllers.TenantReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Tenant")
		os.Exit(1)
	}

	// ---------------------------------------------------------------------
	// NetworkPolicy controller  🔥🔥🔥
	// ---------------------------------------------------------------------
	if err = (&controllers.NetworkPolicyReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkPolicy")
		os.Exit(1)
	}

	// ---------------------------------------------------------------------
	// Health probes
	// ---------------------------------------------------------------------
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
