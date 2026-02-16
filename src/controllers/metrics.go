package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	TenantTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "namespace_operator_tenants_total",
			Help: "Total number of Tenant resources",
		},
	)

	TenantReconcileErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "namespace_operator_reconcile_errors_total",
			Help: "Total number of reconcile errors",
		},
	)

	TenantReconcileDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "namespace_operator_reconcile_duration_seconds",
			Help:    "Reconcile duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	ctrlmetrics.Registry.MustRegister(
		TenantTotal,
		TenantReconcileErrors,
		TenantReconcileDuration,
	)
}
