package controllers

import (
	"os"
	"path/filepath"
	"testing"

	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	testEnv *envtest.Environment
	cfg     *rest.Config
	scheme  = runtime.NewScheme()
)

func TestMain(m *testing.M) {
	// ---------------------------------------------------------------------
	// Register schemes (Go side)
	// ---------------------------------------------------------------------
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme)) // 🔥 ADD THIS
	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))

	// ---------------------------------------------------------------------
	// Start envtest with CRDs
	// ---------------------------------------------------------------------
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "manifests", "charts", "namespace-operator", "crds"), // ⬅️ CRITIQUE
		},
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = testEnv.Stop()
	os.Exit(code)
}
