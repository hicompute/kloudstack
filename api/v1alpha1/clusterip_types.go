package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterIPSpec struct {
	ClusterIPPool string `json:"clusterIPPool"`
	Interface     string `json:"containerInterface"`
	Address       string `json:"address"`
	Mac           string `json:"mac,omitempty"`
	// +kubebuilder:validation:Enum=v4;v6
	Family   string `json:"family"`
	Resource string `json:"resource"`
}

type ClusterIPHistory struct {
	Mac         string      `json:"mac"`
	Interface   string      `json:"interface,omitempty"`
	Resource    string      `json:"resource"`
	AllocatedAt metav1.Time `json:"allocatedAt"`
	ReleasedAt  metav1.Time `json:"releasedAt,omitempty"`
}

type ClusterIPStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	History    []ClusterIPHistory `json:"history,omitempty"`
}

type ClusterIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`
	Spec              ClusterIPSpec   `json:"spec"`
	Status            ClusterIPStatus `json:"status,omitempty,omitzero"`
}

type ClusterIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterIP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterIP{}, &ClusterIPList{})
}
