package controllers

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsIncrement(t *testing.T) {

	// Reset values
	TenantTotal.Set(0)
	TenantReconcileErrors.Add(0)

	// Simulate metrics usage
	TenantTotal.Set(5)
	TenantReconcileErrors.Inc()

	if val := testutil.ToFloat64(TenantTotal); val != 5 {
		t.Fatalf("expected TenantTotal = 5, got %v", val)
	}

	if val := testutil.ToFloat64(TenantReconcileErrors); val != 1 {
		t.Fatalf("expected TenantReconcileErrors = 1, got %v", val)
	}
}
