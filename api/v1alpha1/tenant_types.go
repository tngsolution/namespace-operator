package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TenantSpec struct {
	// Namespace to create/manage
	Namespace string `json:"namespace"`

	// Profile reference (preferred)
	// +optional
	Profile *string `json:"profile,omitempty"`

	// Legacy inline config (deprecated but supported)
	// +optional
	Quota *QuotaSpec `json:"quota,omitempty"`

	// +optional
	Limits *LimitSpec `json:"limits,omitempty"`

	// Network policy rules (ingress/egress)
	// +optional
	Network *NetworkSpec `json:"network,omitempty"`
}

// NetworkSpec définit les règles réseau personnalisées pour un tenant
type NetworkSpec struct {
	// Ingress rules
	// +optional
	Ingress []NetworkPolicyRule `json:"ingress,omitempty"`

	// Egress rules
	// +optional
	Egress []NetworkPolicyRule `json:"egress,omitempty"`
}

// NetworkPolicyRule représente une règle d'ingress ou d'egress simplifiée
type NetworkPolicyRule struct {
	// From définit les sources autorisées (pour ingress)
	// +optional
	From []NetworkPeer `json:"from,omitempty"`

	// To définit les destinations autorisées (pour egress)
	// +optional
	To []NetworkPeer `json:"to,omitempty"`
}

// NetworkPeer représente un peer réseau (podSelector, namespaceSelector, ipBlock)
type NetworkPeer struct {
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// +optional
	IPBlock *IPBlock `json:"ipBlock,omitempty"`
}

// IPBlock permet de spécifier un bloc d'adresses IP
type IPBlock struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
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
