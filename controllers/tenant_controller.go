package controllers

import (
	"context"
	"fmt"
	"strconv"

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

type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// -----------------------------------------------------------------------------
// Configuration resolution
// -----------------------------------------------------------------------------
func (r *TenantReconciler) resolveConfig(
	ctx context.Context,
	tenant *platformv1alpha1.Tenant,
) (*platformv1alpha1.QuotaSpec, *platformv1alpha1.LimitSpec, error) {

	// 1️⃣ TenantProfile takes precedence
	if tenant.Spec.Profile != nil {
		profile := &platformv1alpha1.TenantProfile{}
		if err := r.Get(ctx, client.ObjectKey{
			Name: *tenant.Spec.Profile,
		}, profile); err != nil {
			return nil, nil, err
		}
		return &profile.Spec.Quota, &profile.Spec.Limits, nil
	}

	// 2️⃣ Legacy inline config
	if tenant.Spec.Quota != nil && tenant.Spec.Limits != nil {
		return tenant.Spec.Quota, tenant.Spec.Limits, nil
	}

	return nil, nil, fmt.Errorf(
		"either spec.profile or spec.quota + spec.limits must be set",
	)
}

// -----------------------------------------------------------------------------
// Reconcile
// -----------------------------------------------------------------------------
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var tenant platformv1alpha1.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tenant); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// --------------------------------------------------------------------------
	// Resolve configuration
	// --------------------------------------------------------------------------
	quota, limits, err := r.resolveConfig(ctx, &tenant)
	if err != nil {
		logger.Error(err, "invalid tenant configuration")
		return ctrl.Result{}, nil
	}

	// --------------------------------------------------------------------------
	// Handle deletion (finalizer)
	// --------------------------------------------------------------------------
	if !tenant.DeletionTimestamp.IsZero() {
		ns := &corev1.Namespace{}
		if err := r.Get(ctx, client.ObjectKey{Name: tenant.Spec.Namespace}, ns); err == nil {
			_ = r.Delete(ctx, ns)
		}

		controllerutil.RemoveFinalizer(&tenant, tenantFinalizer)
		_ = r.Update(ctx, &tenant)
		return ctrl.Result{}, nil
	}

	// --------------------------------------------------------------------------
	// Ensure finalizer
	// --------------------------------------------------------------------------
	if !controllerutil.ContainsFinalizer(&tenant, tenantFinalizer) {
		controllerutil.AddFinalizer(&tenant, tenantFinalizer)
		if err := r.Update(ctx, &tenant); err != nil {
			return ctrl.Result{}, err
		}
	}

	nsName := tenant.Spec.Namespace

	// --------------------------------------------------------------------------
	// Namespace
	// --------------------------------------------------------------------------
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
			Labels: map[string]string{
				ManagedByLabelKey: ManagedByLabelValue,
			},
		},
	}

	if err := r.Create(ctx, ns); err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}

	// --------------------------------------------------------------------------
	// ResourceQuota
	// --------------------------------------------------------------------------
	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-quota",
			Namespace: nsName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rq, func() error {
		rq.Spec.Hard = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(quota.CPU),
			corev1.ResourceMemory: resource.MustParse(quota.Memory),
			corev1.ResourcePods:   resource.MustParse(strconv.Itoa(int(quota.Pods))),
		}
		return nil
	}); err != nil {
		return ctrl.Result{}, err
	}

	// --------------------------------------------------------------------------
	// LimitRange
	// --------------------------------------------------------------------------
	lr := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-limits",
			Namespace: nsName,
		},
	}

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
		return ctrl.Result{}, err
	}

	// --------------------------------------------------------------------------
	// Status (PATCH, not UPDATE)
	// --------------------------------------------------------------------------
	original := tenant.DeepCopy()

	setCondition(
		&tenant.Status.Conditions,
		"Ready",
		metav1.ConditionTrue,
		"Reconciled",
		"Namespace, quota and limits applied",
	)

	if err := r.Status().Patch(ctx, &tenant, client.MergeFrom(original)); err != nil {
		logger.Error(err, "unable to patch Tenant status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// -----------------------------------------------------------------------------
// Setup
// -----------------------------------------------------------------------------
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.Tenant{}).
		Complete(r)
}
