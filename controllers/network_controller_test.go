package controllers

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// -----------------------------------------------------------------------------
// Default deny fallback test
// -----------------------------------------------------------------------------
func TestNetworkPolicyReconciler_DefaultDeny(t *testing.T) {
	g := NewWithT(t)

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	reconciler := &NetworkPolicyReconciler{
		Client: k8sClient,
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns",
			Labels: map[string]string{
				ManagedByLabelKey: ManagedByLabelValue,
			},
		},
	}
	g.Expect(k8sClient.Create(context.Background(), ns)).To(Succeed())

	// 🔥 Manually trigger reconcile
	_, err = reconciler.Reconcile(
		context.Background(),
		ctrl.Request{
			NamespacedName: client.ObjectKey{Name: "test-ns"},
		},
	)
	g.Expect(err).NotTo(HaveOccurred())

	// Assert default deny ingress exists
	npIngress := &networkingv1.NetworkPolicy{}
	g.Expect(k8sClient.Get(
		context.Background(),
		client.ObjectKey{Name: "default-deny-ingress", Namespace: "test-ns"},
		npIngress,
	)).To(Succeed())

	// Assert default deny egress exists
	npEgress := &networkingv1.NetworkPolicy{}
	g.Expect(k8sClient.Get(
		context.Background(),
		client.ObjectKey{Name: "default-deny-egress", Namespace: "test-ns"},
		npEgress,
	)).To(Succeed())
}

// -----------------------------------------------------------------------------
// Custom network rules test
// -----------------------------------------------------------------------------
func TestNetworkPolicyReconciler_CustomNetwork(t *testing.T) {
	g := NewWithT(t)

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	g.Expect(err).NotTo(HaveOccurred())

	reconciler := &NetworkPolicyReconciler{
		Client: k8sClient,
	}

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "custom-ns",
			Labels: map[string]string{
				ManagedByLabelKey: ManagedByLabelValue,
			},
		},
	}
	g.Expect(k8sClient.Create(context.Background(), ns)).To(Succeed())

	// Create Tenant with custom network rules
	tenant := &platformv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant1",
		},
		Spec: platformv1alpha1.TenantSpec{
			Namespace: "custom-ns",
			Network: &platformv1alpha1.NetworkSpec{
				Ingress: []platformv1alpha1.NetworkPolicyRule{
					{
						From: []platformv1alpha1.NetworkPeer{
							{
								PodSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "frontend",
									},
								},
							},
						},
					},
				},
				Egress: []platformv1alpha1.NetworkPolicyRule{
					{
						To: []platformv1alpha1.NetworkPeer{
							{
								IPBlock: &platformv1alpha1.IPBlock{
									CIDR: "10.0.0.0/24",
								},
							},
						},
					},
				},
			},
		},
	}

	tenant.SetGroupVersionKind(
		platformv1alpha1.GroupVersion.WithKind("Tenant"),
	)

	g.Expect(k8sClient.Create(context.Background(), tenant)).To(Succeed())

	// 🔥 Manually trigger reconcile
	_, err = reconciler.Reconcile(
		context.Background(),
		ctrl.Request{
			NamespacedName: client.ObjectKey{Name: "custom-ns"},
		},
	)
	g.Expect(err).NotTo(HaveOccurred())

	// Assert custom ingress exists
	npIngress := &networkingv1.NetworkPolicy{}
	g.Expect(k8sClient.Get(
		context.Background(),
		client.ObjectKey{Name: "custom-ingress", Namespace: "custom-ns"},
		npIngress,
	)).To(Succeed())

	// Assert custom egress exists
	npEgress := &networkingv1.NetworkPolicy{}
	g.Expect(k8sClient.Get(
		context.Background(),
		client.ObjectKey{Name: "custom-egress", Namespace: "custom-ns"},
		npEgress,
	)).To(Succeed())
}
