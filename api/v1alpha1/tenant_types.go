package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TenantSpec struct {
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`

	Quota  QuotaSpec `json:"quota"`
	Limits LimitSpec `json:"limits"`
}

type QuotaSpec struct {
	// +kubebuilder:validation:Pattern=`^([0-9]+m|[0-9]+)$`
	CPU string `json:"cpu"`

	// +kubebuilder:validation:Pattern=`^[0-9]+(Mi|Gi)$`
	Memory string `json:"memory"`

	// +kubebuilder:validation:Minimum=1
	Pods int32 `json:"pods"`
}

type LimitSpec struct {
	// +kubebuilder:validation:Pattern=`^([0-9]+m|[0-9]+)$`
	DefaultCPU string `json:"defaultCpu"`

	// +kubebuilder:validation:Pattern=`^[0-9]+(Mi|Gi)$`
	DefaultMemory string `json:"defaultMemory"`

	// +kubebuilder:validation:Pattern=`^([0-9]+m|[0-9]+)$`
	MaxCPU string `json:"maxCpu"`

	// +kubebuilder:validation:Pattern=`^[0-9]+(Mi|Gi)$`
	MaxMemory string `json:"maxMemory"`
}

// +kubebuilder:object:generate=true
type TenantStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=tenant
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
