package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterIPPoolSpec struct {
	// +kubebuilder:validation:Enum=v4;v6
	IPFamily string `json:"ipFamily"`
	CIDR     string `json:"cidr"`
	Gateway  string `json:"gateway,omitempty"`
}

type ClusterIPPoolStatus struct {
	Conditions         []metav1.Condition `json:"conditions"`
	TotalIPs           string             `json:"totalIPs"`
	AllocatedIPs       string             `json:"allocatedIPs"`
	FreeIPs            string             `json:"freeIPs"`
	NextIndex          string             `json:"nextIndex"`
	ReleasedClusterIPs []string           `json:"releasedClusterIPs,omitempty"`
}

type ClusterIPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterIPPoolSpec   `json:"spec"`
	Status            ClusterIPPoolStatus `json:"status"`
}

type ClusterIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterIPPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterIPPool{}, &ClusterIPPoolList{})
}
