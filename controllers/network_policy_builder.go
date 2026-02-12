package controllers

import (
	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildPolicies builds NetworkPolicies for a namespace
func (r *NetworkPolicyReconciler) buildPolicies(
	namespace string,
	tenant *platformv1alpha1.Tenant,
) []*networkingv1.NetworkPolicy {

	// -------------------------------------------------------------------------
	// Custom policies
	// -------------------------------------------------------------------------
	if tenant != nil && tenant.Spec.Network != nil {

		var policies []*networkingv1.NetworkPolicy
		netSpec := tenant.Spec.Network

		// Ingress
		if len(netSpec.Ingress) > 0 {

			np := &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-ingress",
					Namespace: namespace,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
					},
				},
			}

			for _, rule := range netSpec.Ingress {
				var peers []networkingv1.NetworkPolicyPeer

				for _, from := range rule.From {

					peer := networkingv1.NetworkPolicyPeer{
						PodSelector:       from.PodSelector,
						NamespaceSelector: from.NamespaceSelector,
					}

					if from.IPBlock != nil {
						peer.IPBlock = &networkingv1.IPBlock{
							CIDR:   from.IPBlock.CIDR,
							Except: from.IPBlock.Except,
						}
					}

					peers = append(peers, peer)
				}

				np.Spec.Ingress = append(
					np.Spec.Ingress,
					networkingv1.NetworkPolicyIngressRule{
						From: peers,
					},
				)
			}

			policies = append(policies, np)
		}

		// Egress
		if len(netSpec.Egress) > 0 {

			np := &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-egress",
					Namespace: namespace,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeEgress,
					},
				},
			}

			for _, rule := range netSpec.Egress {
				var peers []networkingv1.NetworkPolicyPeer

				for _, to := range rule.To {

					peer := networkingv1.NetworkPolicyPeer{
						PodSelector:       to.PodSelector,
						NamespaceSelector: to.NamespaceSelector,
					}

					if to.IPBlock != nil {
						peer.IPBlock = &networkingv1.IPBlock{
							CIDR:   to.IPBlock.CIDR,
							Except: to.IPBlock.Except,
						}
					}

					peers = append(peers, peer)
				}

				np.Spec.Egress = append(
					np.Spec.Egress,
					networkingv1.NetworkPolicyEgressRule{
						To: peers,
					},
				)
			}

			policies = append(policies, np)
		}

		if len(policies) > 0 {
			return policies
		}
	}

	// -------------------------------------------------------------------------
	// Default deny fallback
	// -------------------------------------------------------------------------
	return []*networkingv1.NetworkPolicy{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default-deny-ingress",
				Namespace: namespace,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default-deny-egress",
				Namespace: namespace,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeEgress,
				},
			},
		},
	}
}
