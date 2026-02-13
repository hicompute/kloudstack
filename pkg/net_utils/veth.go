package netutils

import (
	"net"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

func SetupVeth(contNetnsPath, contIfaceName, requestedMac string, mtu int) (*current.Interface, *current.Interface, error) {
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
