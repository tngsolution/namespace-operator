package controllers

import (
	"context"

	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="platform.example.com",resources=tenants,verbs=get;list;watch
// +kubebuilder:rbac:groups="networking.k8s.io",resources=networkpolicies,verbs=get;list;watch;create;update;patch

type NetworkPolicyReconciler struct {
	client.Client
}

// -----------------------------------------------------------------------------

func (r *NetworkPolicyReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	// -------------------------------------------------------------------------
	// Get Namespace
	// -------------------------------------------------------------------------
	var ns corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &ns); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// -------------------------------------------------------------------------
	// Find Tenant owning this namespace (envtest-safe)
	// -------------------------------------------------------------------------
	var tenants platformv1alpha1.TenantList
	if err := r.List(ctx, &tenants); err != nil {
		logger.Error(err, "unable to list Tenants")
		return ctrl.Result{}, err
	}

	var tenant *platformv1alpha1.Tenant
	for i := range tenants.Items {
		if tenants.Items[i].Spec.Namespace == ns.Name {
			tenant = &tenants.Items[i]
			break
		}
	}

	// -------------------------------------------------------------------------
	// Build policies (external builder file)
	// -------------------------------------------------------------------------
	policies := r.buildPolicies(ns.Name, tenant)

	// -------------------------------------------------------------------------
	// Apply policies (Server-Side Apply requires GVK!)
	// -------------------------------------------------------------------------
	for _, np := range policies {

		// 🔥 REQUIRED for SSA + envtest
		np.SetGroupVersionKind(
			networkingv1.SchemeGroupVersion.WithKind("NetworkPolicy"),
		)

		if err := r.Patch(
			ctx,
			np,
			client.Apply,
			client.FieldOwner("network-controller"),
			client.ForceOwnership,
		); err != nil && !apierrors.IsAlreadyExists(err) {

			logger.Error(err, "unable to apply NetworkPolicy", "name", np.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// -----------------------------------------------------------------------------

func (r *NetworkPolicyReconciler) SetupWithManager(
	mgr ctrl.Manager,
) error {

	// Index Tenant by spec.namespace
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&platformv1alpha1.Tenant{},
		"spec.namespace",
		func(obj client.Object) []string {
			tenant := obj.(*platformv1alpha1.Tenant)
			if tenant.Spec.Namespace == "" {
				return nil
			}
			return []string{tenant.Spec.Namespace}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Watches(
			&platformv1alpha1.Tenant{},
			handler.EnqueueRequestsFromMapFunc(
				func(ctx context.Context, obj client.Object) []reconcile.Request {

					tenant, ok := obj.(*platformv1alpha1.Tenant)
					if !ok || tenant.Spec.Namespace == "" {
						return nil
					}

					return []reconcile.Request{
						{
							NamespacedName: client.ObjectKey{
								Name: tenant.Spec.Namespace,
							},
						},
					}
				},
			),
		).
		Complete(r)
}
