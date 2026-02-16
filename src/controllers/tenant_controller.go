package controllers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	platformv1alpha1 "github.com/tngs/namespace-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const tenantFinalizer = "platform.example.com/finalizer"

// -----------------------------------------------------------------------------
// RBAC
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=platform.example.com,resources=tenants,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=platform.example.com,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=resourcequotas,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=limitranges,verbs=get;list;watch;create;update;patch

type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// ---------------------------------------------------------------------------------------------------
// The resolveConfig function determines the effective quota and limits for a Tenant.
// It checks if a profile is specified and fetches the corresponding quota and limits.
// If no profile is specified, it expects both quota and limits to be set directly on the Tenant.
// This function ensures that the controller can support both profile-based and direct configuration.
// ----------------------------------------------------------------------------------------------------
func (r *TenantReconciler) resolveConfig(
	ctx context.Context,
	tenant *platformv1alpha1.Tenant,
) (*platformv1alpha1.QuotaSpec, *platformv1alpha1.LimitSpec, error) {

	// If a profile is specified, it takes precedence over direct quota/limit settings
	if tenant.Spec.Profile != nil {
		// Fetch the profile to get the quota and limits
		profile := &platformv1alpha1.TenantProfile{}
		if err := r.Get(ctx, client.ObjectKey{
			Name: *tenant.Spec.Profile,
		}, profile); err != nil {
			return nil, nil, err
		}

		return &profile.Spec.Quota, &profile.Spec.Limits, nil
	}

	// If no profile is specified, both quota and limits must be set directly on the tenant
	if tenant.Spec.Quota != nil && tenant.Spec.Limits != nil {
		return tenant.Spec.Quota, tenant.Spec.Limits, nil
	}

	return nil, nil, fmt.Errorf(
		"either spec.profile or spec.quota + spec.limits must be set",
	)
}

// -----------------------------------------------------------------------------
// Reconcile
// The Reconcile function is the heart of the controller.
// It is called whenever a Tenant resource is created, updated, or deleted.
// -----------------------------------------------------------------------------
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	start := time.Now()
	defer func() {
		TenantReconcileDuration.Observe(time.Since(start).Seconds())
	}()

	logger := log.FromContext(ctx)

	var tenant platformv1alpha1.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tenant); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// -------------------------------------------------------------------------
	// Deletion handling (finalizer)
	// -------------------------------------------------------------------------
	if !tenant.DeletionTimestamp.IsZero() {

		ns := &corev1.Namespace{}
		if err := r.Get(ctx, client.ObjectKey{Name: tenant.Spec.Namespace}, ns); err == nil {
			_ = r.Delete(ctx, ns)
		}

		controllerutil.RemoveFinalizer(&tenant, tenantFinalizer)
		if err := r.Update(ctx, &tenant); err != nil {
			TenantReconcileErrors.Inc()
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// ------------------------------------------------------------------------------------------------------------------------------------------------
	// If the tenant is not being deleted, ensure it has the finalizer so we can clean up resources on deletion
	// This is important to prevent orphaned namespaces if a tenant is deleted without the finalizer
	// The finalizer allows us to perform cleanup (like deleting the namespace) before the tenant resource is actually removed from the cluster
	// If the tenant doesn't have the finalizer, add it and update the resource. This will trigger another reconcile loop, but that's fine.
	// We only add the finalizer if it's not already present, to avoid unnecessary updates.
	// This pattern is common in Kubernetes controllers to manage the lifecycle of related resources.
	// ------------------------------------------------------------------------------------------------------------------------------------------------
	if !controllerutil.ContainsFinalizer(&tenant, tenantFinalizer) {
		controllerutil.AddFinalizer(&tenant, tenantFinalizer)
		if err := r.Update(ctx, &tenant); err != nil {
			TenantReconcileErrors.Inc()
			return ctrl.Result{}, err
		}
	}

	// -------------------------------------------------------------------------
	// Resolve configuration
	// -------------------------------------------------------------------------e.
	quota, limits, err := r.resolveConfig(ctx, &tenant)
	if err != nil {
		logger.Error(err, "invalid tenant configuration")
		TenantReconcileErrors.Inc()
		return ctrl.Result{}, nil
	}

	nsName := tenant.Spec.Namespace

	// -------------------------------------------------------------------------
	// Namespace
	// -------------------------------------------------------------------------
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
			Labels: map[string]string{
				ManagedByLabelKey: ManagedByLabelValue,
			},
		},
	}

	// Set the tenant as the owner of the namespace, so it gets deleted automatically when the tenant is deleted
	if err := controllerutil.SetControllerReference(&tenant, ns, r.Scheme); err != nil {
		TenantReconcileErrors.Inc()
		return ctrl.Result{}, err
	}

	// Create the namespace if it doesn't exist. If it already exists, that's fine -
	// we just want to ensure it exists and is owned by the tenant.
	if err := r.Create(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
		TenantReconcileErrors.Inc()
		return ctrl.Result{}, err
	}

	// -------------------------------------------------------------------------
	// ResourceQuota
	// -------------------------------------------------------------------------
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-quota",
			Namespace: nsName,
		},
	}

	// Create or update the ResourceQuota with the specified limits.
	// If it doesn't exist, it will be created. If it already exists,
	// it will be updated with the new limits.
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rq, func() error {
		rq.Spec.Hard = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(quota.CPU),
			corev1.ResourceMemory: resource.MustParse(quota.Memory),
			corev1.ResourcePods:   resource.MustParse(strconv.Itoa(int(quota.Pods))),
		}
		return nil
	}); err != nil {
		TenantReconcileErrors.Inc()
		return ctrl.Result{}, err
	}

	// -------------------------------------------------------------------------
	// LimitRange
	// -------------------------------------------------------------------------
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-limits",
			Namespace: nsName,
		},
	}

	// Create or update the LimitRange with the specified limits.
	// If it doesn't exist, it will be created. If it already exists,
	// it will be updated with the new limits.
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, lr, func() error {
		lr.Spec.Limits = []corev1.LimitRangeItem{
			{
				Type: corev1.LimitTypeContainer,
				Default: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(limits.DefaultCPU),
					corev1.ResourceMemory: resource.MustParse(limits.DefaultMemory),
				},
				Max: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(limits.MaxCPU),
					corev1.ResourceMemory: resource.MustParse(limits.MaxMemory),
				},
			},
		}
		return nil
	}); err != nil {
		TenantReconcileErrors.Inc()
		return ctrl.Result{}, err
	}

	// -------------------------------------------------------------------------
	// Update metrics (total tenants)
	// -------------------------------------------------------------------------
	var tenantList platformv1alpha1.TenantList
	if err := r.List(ctx, &tenantList); err == nil {
		TenantTotal.Set(float64(len(tenantList.Items)))
	}

	// -------------------------------------------------------------------------
	// Status
	// -------------------------------------------------------------------------
	original := tenant.DeepCopy()

	setCondition(
		&tenant.Status.Conditions,
		"Ready",
		metav1.ConditionTrue,
		"Reconciled",
		"Namespace, quota and limits applied",
	)

	// Patch the status subresource to update the status of the tenant.
	// We use Patch with MergeFrom to avoid overwriting any changes that might have been made to the tenant resource since we fetched it.
	if err := r.Status().Patch(ctx, &tenant, client.MergeFrom(original)); err != nil {
		TenantReconcileErrors.Inc()
		logger.Error(err, "unable to patch Tenant status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// -----------------------------------------------------------------------------
// The SetupWithManager function sets up the controller with the Manager.
// It tells the controller to watch for Tenant resources and also to watch for
// Namespaces that are owned by Tenants, so it can react to changes in those as well.
// -----------------------------------------------------------------------------
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.Tenant{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
