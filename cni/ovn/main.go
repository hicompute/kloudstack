package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	ovnCniTypes "github.com/hicompute/kloudstack/pkg/daemons/ovn-cni-server/types"
)

func main() {
	runtime.LockOSThread()
	funcs := skel.CNIFuncs{
		Add: cmdAdd,
		Del: cmdDel,
	}
	skel.PluginMainFuncs(funcs, version.All, "kloudstack-ovn")
}

func cmdAdd(args *skel.CmdArgs) error {
	result, err := runOnDaemon("Add", args)
	if err != nil {
		return err
	}
	return types.PrintResult(result, version.Current())
}

func cmdDel(args *skel.CmdArgs) error {
	_, err := runOnDaemon("Del", args)
	if err != nil {
		return err
	}
	return nil
}

func runOnDaemon(CMD string, args *skel.CmdArgs) (*current.Result, error) {
	cniDsocket := "/var/run/kloudstack-ovn-cni.sock"
	var resp ovnCniTypes.CNIResponse
	if err := callCniDaemon(cniDsocket, ovnCniTypes.CNIRequest{
		Cmd:     CMD,
		CmdArgs: *args,
	}, &resp); err != nil {
		return nil, fmt.Errorf("kloudstack ovn cni daemon request failed: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	return &resp.Result, nil
}

func callCniDaemon(socket string, req ovnCniTypes.CNIRequest, resp *ovnCniTypes.CNIResponse) error {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return err
	}
	defer conn.Close()

	byteData, _ := json.Marshal(req)
	if _, err := conn.Write(byteData); err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(conn)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf.Bytes(), &resp)
}
