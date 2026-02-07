package dhcp

import (
	"net"
	"strings"

	kloudstack_ipam "github.com/hicompute/kloudstack/pkg/ipam"
	"github.com/spf13/viper"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/samber/lo"
	"k8s.io/klog/v2"
)

type HDHCPV4 struct {
	ipam   kloudstack_ipam.IPAM
	server server4.Server
}

func StartV4Server(ifaceName string) error {
	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: 67}

	ipam := kloudstack_ipam.New()
	h := &HDHCPV4{ipam: *ipam}

	srv, err := server4.NewServer(ifaceName, laddr, h.handler, server4.WithDebugLogger())
	if err != nil {
		return err
	}
	return srv.Serve()
}

func (hd4 *HDHCPV4) handler(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	macPrefix := viper.GetString("MACPREFIX")
	mac := req.ClientHWAddr.String()
	if !strings.HasPrefix(mac, macPrefix) {
		return
	}

	// Log basic info
	klog.Infof("=== DHCP %s from %s ===", req.MessageType(), req.ClientHWAddr)
	klog.Infof("XID: %s CIAddr: %s SIAddr: %s", req.TransactionID, req.ClientIPAddr, req.ServerIPAddr)

	switch req.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		hd4.handleDiscover(conn, peer, req)
	case dhcpv4.MessageTypeRequest:
		hd4.handleRequest(conn, peer, req)
	default:
		klog.Infof("Ignoring DHCP message: %s", req.MessageType())
	}
}

// ------------------------------------------------------------
//  SHARED HELPERS (DRY)
// ------------------------------------------------------------

// buildReply creates a reply packet (Offer or ACK)
func buildReply(req *dhcpv4.DHCPv4, msgType dhcpv4.MessageType) (*dhcpv4.DHCPv4, error) {
	reply, err := dhcpv4.NewReplyFromRequest(req)
	if err != nil {
		return nil, err
	}

	reply.UpdateOption(dhcpv4.OptMessageType(msgType))

	return reply, nil
}

// applyCommonOptions sets mask, server ID, routes, etc.
func (hd4 *HDHCPV4) applyCommonOptions(pkt *dhcpv4.DHCPv4) {
	clusterIP, err := hd4.ipam.FindClusterIPbyFamilyandMAC(pkt.ClientHWAddr.String(), "v4")
	if err != nil {
		klog.Errorf("IPAM lookup failed: %v", err)
		return
	}
	clusterIPPool, err := hd4.ipam.FindClusterIPPoolByName(clusterIP.Spec.ClusterIPPool)
	if err != nil {
		klog.Errorf("%v", err)
		return
	}
	_, ipNet, _ := net.ParseCIDR(clusterIPPool.Spec.CIDR)
	pkt.UpdateOption(dhcpv4.OptSubnetMask(ipNet.Mask))
	ip := net.ParseIP(clusterIP.Spec.Address)
	pkt.YourIPAddr = ip

	if clusterIPPool.Spec.Gateway != "" {
		var routes []*dhcpv4.Route
		gwIP := net.ParseIP(clusterIPPool.Spec.Gateway)
		if !ipNet.Contains(gwIP) {
			p2pRoute := &dhcpv4.Route{
				Dest: &net.IPNet{
					IP:   gwIP,
					Mask: net.CIDRMask(32, 32),
				},
				Router: net.ParseIP("0.0.0.0"),
			}
			routes = append(routes, p2pRoute)
		}
		defaultRoute := &dhcpv4.Route{
			Dest: &net.IPNet{
				IP:   net.ParseIP("0.0.0.0"),
				Mask: net.CIDRMask(0, 32),
			},
			Router: gwIP,
		}
		routes = append(routes, defaultRoute)
		pkt.UpdateOption(dhcpv4.OptClasslessStaticRoute(routes...))
	}

	// Server IP
	serverIP := net.ParseIP(viper.GetString("SERVER"))
	pkt.ServerIPAddr = serverIP
	pkt.UpdateOption(dhcpv4.OptServerIdentifier(serverIP))

	dnsServers := lo.Map(strings.Split(viper.GetString("DNS"), ","), func(d string, index int) net.IP {
		return net.ParseIP(d)
	})

	pkt.UpdateOption(dhcpv4.OptDNS(dnsServers...))

}

// sendPacket writes the DHCP packet
func sendPacket(conn net.PacketConn, peer net.Addr, pkt *dhcpv4.DHCPv4) {
	_, err := conn.WriteTo(pkt.ToBytes(), peer)
	if err != nil {
		klog.Errorf("send failed: %v", err)
	}
}

// ------------------------------------------------------------
//              OFFER HANDLER
// ------------------------------------------------------------

func (hd4 *HDHCPV4) handleDiscover(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {

	offer, err := buildReply(req, dhcpv4.MessageTypeOffer)
	if err != nil {
		klog.Errorf("Offer reply build failed: %v", err)
		return
	}

	hd4.applyCommonOptions(offer)

	klog.Infof("Sending OFFER → %s", offer.YourIPAddr)

	sendPacket(conn, peer, offer)
}

// ------------------------------------------------------------
//              REQUEST HANDLER
// ------------------------------------------------------------

func (hd4 *HDHCPV4) handleRequest(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {

	ip := req.RequestedIPAddress()
	if ip == nil {
		klog.Errorf("Client REQUEST had no RequestedIPAddress")
		return
	}

	ack, err := buildReply(req, dhcpv4.MessageTypeAck)
	if err != nil {
		klog.Errorf("Ack reply build failed: %v", err)
		return
	}

	ack.YourIPAddr = ip

	hd4.applyCommonOptions(ack)

	klog.Infof("Sending ACK → %s", ack.YourIPAddr)

	sendPacket(conn, peer, ack)
}
