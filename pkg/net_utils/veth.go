package netutils

import (
	"net"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

func SetupVeth(contNetnsPath, contIfaceName, requestedMac string, mtu int, ipAddress *net.IPNet, gatewayAddress *net.IP) (*current.Interface, *current.Interface, error) {
	hostIface := &current.Interface{}
	contIface := &current.Interface{}
	contNetns, err := ns.GetNS(contNetnsPath)
	if err != nil {
		return nil, nil, err
	}
	defer contNetns.Close()

	err = contNetns.Do(func(hostNetns ns.NetNS) error {
		hostVeth, containerVeth, err := ip.SetupVeth(contIfaceName, mtu, requestedMac, hostNetns)
		if err != nil {
			return err
		}
		if err := setInterfaceUp(contIfaceName); err != nil {
			return err
		}
		klog.Infof("ip address: %v, gateway: %v", ipAddress, *gatewayAddress)
		// if ipAddress != nil {
		// 	link, err := AddInterfaceIPAddress(contIfaceName, &netlink.Addr{
		// 		IPNet:     ipAddress,
		// 		LinkIndex: containerVeth.Index,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
		// 	klog.Infof("ip address: %v", ipAddress)
		// 	if gatewayAddress != nil {
		// 		klog.Infof("gateay ip address: %v", *gatewayAddress)
		// 		dr := netlink.Route{
		// 			Dst:       nil,
		// 			Gw:        *gatewayAddress,
		// 			Flags:     int(netlink.FLAG_ONLINK),
		// 			LinkIndex: link.Attrs().Index,
		// 			// Scope:     netlink.SCOPE_NOWHERE,
		// 		}
		// 		if err := netlink.RouteReplace(&dr); err != nil {
		// 			klog.Errorf("error on route add: %v", err)
		// 			return err
		// 		}
		// 	}
		// }

		contIface.Name = containerVeth.Name
		contIface.Mac = containerVeth.HardwareAddr.String()
		contIface.Sandbox = contNetns.Path()
		hostIface.Name = hostVeth.Name
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Refetch the hostIface since its MAC address may change during network namespace move.
	if err = refetchIface(hostIface); err != nil {
		return nil, nil, err
	}

	return hostIface, contIface, nil
}

func setInterfaceUp(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}

	return nil
}

func refetchIface(iface *current.Interface) error {
	macAddress, err := getHardwareAddr(iface.Name)
	if err != nil {
		return err
	}
	iface.Mac = *macAddress
	return nil
}

func getHardwareAddr(ifName string) (*string, error) {
	ifLink, err := netlink.LinkByName(ifName)
	if err != nil {
		return nil, err
	}
	macAddress := ifLink.Attrs().HardwareAddr.String()
	return &macAddress, nil
}

func setHardwareAddr(ifName, macAddress string) error {
	ifLink, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetHardwareAddr(ifLink, net.HardwareAddr(macAddress)); err != nil {
		return err
	}
	return nil
}

func AddInterfaceIPAddress(ifName string, ipAddress *netlink.Addr) (netlink.Link, error) {
	ifLink, err := netlink.LinkByName(ifName)
	if err != nil {
		return nil, err
	}
	if err := netlink.AddrAdd(ifLink, ipAddress); err != nil {
		return nil, err
	}
	return ifLink, nil
}
