package controllers

import (
	"context"
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
	// Handle deletion with finalizer
	// --------------------------------------------------------------------------
	if !tenant.DeletionTimestamp.IsZero() {
		ns := &corev1.Namespace{}
		if err := r.Get(ctx, client.ObjectKey{Name: tenant.Spec.Namespace}, ns); err == nil {
			_ = r.Delete(ctx, ns) // policy: delete namespace
		}

		controllerutil.RemoveFinalizer(&tenant, tenantFinalizer)
		if err := r.Update(ctx, &tenant); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// --------------------------------------------------------------------------
	// Ensure finalizer is present
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

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, rq, func() error {
		rq.Spec.Hard = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(tenant.Spec.Quota.CPU),
			corev1.ResourceMemory: resource.MustParse(tenant.Spec.Quota.Memory),
			corev1.ResourcePods:   resource.MustParse(strconv.Itoa(int(tenant.Spec.Quota.Pods))),
		}
		return nil
	})
	if err != nil {
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

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, lr, func() error {
		lr.Spec.Limits = []corev1.LimitRangeItem{
			{
				Type: corev1.LimitTypeContainer,
				Default: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(tenant.Spec.Limits.DefaultCPU),
					corev1.ResourceMemory: resource.MustParse(tenant.Spec.Limits.DefaultMemory),
				},
				Max: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(tenant.Spec.Limits.MaxCPU),
					corev1.ResourceMemory: resource.MustParse(tenant.Spec.Limits.MaxMemory),
				},
			},
		}
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// --------------------------------------------------------------------------
	// Status
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

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.Tenant{}).
		Complete(r)
}
