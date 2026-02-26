package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterIPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterIPPoolSpec   `json:"spec"`
	Status            ClusterIPPoolStatus `json:"status"`
}

type ClusterIPPoolSpec struct {
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterIPPool `json:"items"`
}
