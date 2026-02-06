package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=tp
// +kubebuilder:subresource:status
type TenantProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TenantProfileSpec `json:"spec,omitempty"`
}

type TenantProfileSpec struct {
	Quota  QuotaSpec `json:"quota"`
	Limits LimitSpec `json:"limits"`
}

// +kubebuilder:object:root=true
type TenantProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantProfile{}, &TenantProfileList{})
}
