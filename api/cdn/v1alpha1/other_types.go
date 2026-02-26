package v1alpha1

type Group string
type Kind string
type Namespace string
type ObjectName string

type LocalObjectReference struct {
	Group Group      `json:"group"`
	Kind  Kind       `json:"kind"`
	Name  ObjectName `json:"name"`
}

type SecretObjectReference struct {
	Group     *Group     `json:"group"`
	Kind      *Kind      `json:"kind"`
	Name      ObjectName `json:"name"`
	Namespace *Namespace `json:"namespace,omitempty"`
}

type Path struct {
	Type string `json:"type,omitempty"`
	Path string `json:"path"`
}
