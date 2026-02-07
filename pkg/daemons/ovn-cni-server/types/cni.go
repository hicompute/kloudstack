package types

import (
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
)

type CNIRequest struct {
	Cmd     string       `json:"cmd"`
	CmdArgs skel.CmdArgs `jsob:"cmdArgs"`
}

type CNIResponse struct {
	Result current.Result `json:"result"`
	Error  string         `json:"error,omitempty"`
}

type CniKubeArgs struct {
	types.CommonArgs
	K8S_POD_NAME               types.UnmarshallableString
	K8S_POD_NAMESPACE          types.UnmarshallableString
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString
	K8S_POD_UID                types.UnmarshallableString
}
