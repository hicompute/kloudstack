package v1alpha1

func Init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
	SchemeBuilder.Register(&HTTPRoute{}, &HTTPRouteList{})
	SchemeBuilder.Register(&WAFRule{})
}
