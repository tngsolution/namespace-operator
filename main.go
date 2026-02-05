package main

import (
	"os"

	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"
	"github.com/tngs/namespace-operator/controllers"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))           // ✅ Namespace, Pod, etc
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme)) // ✅ Tenant CRD
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})

	if err != nil {
		os.Exit(1)
	}

	if err = (&controllers.TenantReconciler{
		Client: mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
