package ipam

type IPAMRequest struct {
	Namespace string  `json:"namespace"`
	Name      string  `json:"name"`
	Interface string  `json:"interface"`
	Mac       *string `json:"mac,omitempty"`
	Family    string  `json:"family"`
}
