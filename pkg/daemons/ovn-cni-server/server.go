package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	types100 "github.com/containernetworking/cni/pkg/types/100"

	"github.com/containernetworking/cni/pkg/version"
	cniTypes "github.com/hicompute/kloudstack/pkg/daemons/ovn-cni-server/types"
	helper "github.com/hicompute/kloudstack/pkg/helpers"
	kloudstack_ipam "github.com/hicompute/kloudstack/pkg/ipam"
	netUtils "github.com/hicompute/kloudstack/pkg/net_utils"
	netutils "github.com/hicompute/kloudstack/pkg/net_utils"
	"github.com/hicompute/kloudstack/pkg/ovn"
	"github.com/hicompute/kloudstack/pkg/ovs"
	"k8s.io/klog/v2"
)

type CNIServer struct {
	socketPath string
	listener   net.Listener
	ovsAgent   ovs.OvsAgent
	ovnAgent   ovn.OVNagent
	ipam       kloudstack_ipam.IPAM
}

func Start(socketPath string) error {
	// Cleanup existing socket
	os.RemoveAll(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket: %v", err)
	}

	ovsAgent, err := ovs.CreateOVSagent()
	if err != nil {
		return fmt.Errorf("failed to create ovs agent: %v", err)
	}

	ovnAgent, err := ovn.CreateOVNagent("tcp:192.168.12.177:6641")
	if err != nil {
		return fmt.Errorf("failed to create ovs agent: %v", err)
	}

	ipam := kloudstack_ipam.New()

	cniServer := &CNIServer{
		socketPath: socketPath,
		listener:   listener,
		ovsAgent:   *ovsAgent,
		ovnAgent:   *ovnAgent,
		ipam:       *ipam,
	}

	cniServer.run()
	return nil
}

func (s *CNIServer) run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			klog.Errorf("Failed to accept CNI connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *CNIServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	var request cniTypes.CNIRequest
	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		klog.Errorf("Failed to accept CNI connection: %v", err)
		return
	}

	var response cniTypes.CNIResponse
	switch request.Cmd {
	case "Add":
		response = s.handleAdd(request.CmdArgs)
	case "Del":
		response = s.handleDel(request.CmdArgs)
	// case "CHECK":
	// 	response = s.handleCheck(request)
	default:
		response = cniTypes.CNIResponse{Error: "Unknown command"}
	}

	json.NewEncoder(conn).Encode(response)
}

func (s *CNIServer) handleAdd(req skel.CmdArgs) cniTypes.CNIResponse {

	conf, err := parseConfig(req.StdinData)
	if err != nil {
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	klog.Infof("logical switch: %s", conf.LogicalSwitch)

	k8sArgs := cniTypes.CniKubeArgs{}
	if err := types.LoadArgs(req.Args, &k8sArgs); err != nil {
		klog.Infof("error loading args: %v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	K8S_POD_NAMESPACE := string(k8sArgs.K8S_POD_NAMESPACE)
	K8S_POD_NAME := string(k8sArgs.K8S_POD_NAME)
	vmName := helper.ExtractVMName(K8S_POD_NAME)

	var ipAddress net.IPNet
	var mac string

	ls := conf.LogicalSwitch
	if ls == "" {
		ls = "public"

		clusterIP, clusterIPPool, err := s.ipam.FindOrCreateClusterIP(kloudstack_ipam.IPAMRequest{
			Interface: req.IfName,
			Namespace: K8S_POD_NAMESPACE,
			Name:      K8S_POD_NAME,
			Family:    "v4",
		})

		if err != nil {
			return cniTypes.CNIResponse{
				Error: err.Error(),
			}
		}

		_, ipNet, err := net.ParseCIDR(clusterIPPool.Spec.CIDR)
		ipAddress = net.IPNet{IP: net.ParseIP(clusterIP.Spec.Address), Mask: net.IPMask(ipNet.Mask)}
		mac = clusterIP.Spec.Mac
	} else {
		ls = K8S_POD_NAMESPACE + "/" + ls
		mac = netutils.GenerateVethMAC(vmName, "02")
		ipAddress = net.IPNet{IP: net.ParseIP("192.168.1.1"), Mask: net.IPMask(net.CIDRMask(24, 32))}
	}
	hostIface, contIface, err := netUtils.SetupVeth(req.Netns, req.IfName, mac, 1500)
	if err != nil {
		klog.Errorf("%v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	klog.Info(hostIface.Mac, ",", contIface.Mac)

	ifaceId := K8S_POD_NAMESPACE + "_" + K8S_POD_NAME + "_" + req.IfName

	if err = s.ovsAgent.AddPort("br-int", hostIface.Name, "system", ifaceId); err != nil {
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}

	if err := s.ovnAgent.CreateLogicalPort(ls, ifaceId, contIface.Mac, map[string]string{
		"namespace": K8S_POD_NAMESPACE,
		"pod":       K8S_POD_NAME,
		"vmName":    vmName,
	}); err != nil {
		_ = s.ovsAgent.DelPort("br-int", ifaceId)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	result := current.Result{
		CNIVersion: version.Current(),
		Interfaces: []*current.Interface{contIface},
	}

	if req.IfName == "eth0" {
		result.IPs = []*current.IPConfig{
			{
				Interface: types100.Int(0),
				Address:   ipAddress,
				// Gateway:   gateway,
			},
		}
	}
	return cniTypes.CNIResponse{
		Result: result,
		Error:  "",
	}
}

func (s *CNIServer) handleDel(req skel.CmdArgs) cniTypes.CNIResponse {
	k8sArgs := cniTypes.CniKubeArgs{}
	if err := types.LoadArgs(req.Args, &k8sArgs); err != nil {
		klog.Infof("error loading args: %v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	K8S_POD_NAMESPACE := string(k8sArgs.K8S_POD_NAMESPACE)
	K8S_POD_NAME := string(k8sArgs.K8S_POD_NAME)
	ifaceId := K8S_POD_NAMESPACE + "_" + K8S_POD_NAME + "_" + req.IfName

	if err := s.ovnAgent.DeleteLogicalPort("public", ifaceId); err != nil {
		klog.Errorf("%v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	if err := s.ovsAgent.DelPort("br-int", ifaceId); err != nil {
		klog.Errorf("%v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	return cniTypes.CNIResponse{}
}

func parseConfig(stdin []byte) (*cniTypes.PluginConf, error) {
	conf := cniTypes.PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	if err := version.ParsePrevResult(&conf.NetConf); err != nil {
		return nil, fmt.Errorf("could not parse prevResult: %v", err)
	}

	return &conf, nil
}
