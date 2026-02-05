package controllers

import (
	"context"
	"testing"
	"time"

	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestTenantCreatesNamespace(t *testing.T) {
	g := NewWithT(t)

	// -------------------------------------------------------------------------
	// Client EXPLICITE avec le bon Scheme
	// -------------------------------------------------------------------------
	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	// -------------------------------------------------------------------------
	// Manager
	// -------------------------------------------------------------------------
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	reconciler := &TenantReconciler{
		Client: k8sClient,
	}
	g.Expect(reconciler.SetupWithManager(mgr)).To(Succeed())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = mgr.Start(ctx)
	}()

	// -------------------------------------------------------------------------
	// Tenant CR (GVK FORCÉ)
	// -------------------------------------------------------------------------
	tenant := &platformv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "team-a",
		},
		Spec: platformv1alpha1.TenantSpec{
			Namespace: "team-a",
			Quota: platformv1alpha1.QuotaSpec{
				CPU:    "1",
				Memory: "1Gi",
				Pods:   5,
			},
			Limits: platformv1alpha1.LimitSpec{
				DefaultCPU:    "100m",
				DefaultMemory: "128Mi",
				MaxCPU:        "500m",
				MaxMemory:     "512Mi",
			},
		},
	}

	// 🔑 LIGNE CRITIQUE : forcer le GVK
	tenant.SetGroupVersionKind(
		platformv1alpha1.GroupVersion.WithKind("Tenant"),
	)

	g.Expect(k8sClient.Create(context.Background(), tenant)).To(Succeed())

	// -------------------------------------------------------------------------
	// Vérifier que le Namespace est créé
	// -------------------------------------------------------------------------
	ns := &corev1.Namespace{}
	g.Eventually(func() error {
		return k8sClient.Get(
			context.Background(),
			client.ObjectKey{Name: "team-a"},
			ns,
		)
	}, 10*time.Second).Should(Succeed())
}
