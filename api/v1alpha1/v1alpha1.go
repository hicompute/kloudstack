package v1alpha1

func Init() {
	SchemeBuilder.Register(&ClusterIPPool{}, &ClusterIPPoolList{})
	SchemeBuilder.Register(&ClusterIP{}, &ClusterIPList{})
}
