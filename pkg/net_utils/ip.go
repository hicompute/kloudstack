package netutils

import (
	"fmt"
	"math/big"
	"net"
)

func PickUsableIPFromCIDRIndex(cidr string, index *big.Int) (string, error) {
	if index == nil {
		return "", fmt.Errorf("index cannot be nil")
	}

	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	ones, bits := ipNet.Mask.Size()
	hostBits := bits - ones
	if hostBits < 0 {
		return "", fmt.Errorf("invalid mask for %s", cidr)
	}

	total := new(big.Int).Lsh(big.NewInt(1), uint(hostBits))
	isIPv4 := ip.To4() != nil

	var offset *big.Int

	if isIPv4 {
		switch hostBits {
		case 0:
			// /32 → exactly one usable address
			if index.Sign() != 0 {
				return "", fmt.Errorf("index %s out of range for %s", index, cidr)
			}
			offset = big.NewInt(0)

		case 1:
			// /31 → two usable addresses (RFC 3021)
			if index.Sign() < 0 || index.Cmp(big.NewInt(1)) > 0 {
				return "", fmt.Errorf("index %s out of range for %s", index, cidr)
			}
			offset = new(big.Int).Set(index)

		default:
			// /30+ → exclude network & broadcast
			max := new(big.Int).Sub(total, big.NewInt(2)) // usable count
			if index.Sign() < 0 || index.Cmp(max) > 0 {
				return "", fmt.Errorf("index %s out of usable range for %s", index, cidr)
			}
			offset = new(big.Int).Add(index, big.NewInt(1))
		}
	} else {
		// IPv6: all addresses usable
		if index.Sign() < 0 || index.Cmp(total) >= 0 {
			return "", fmt.Errorf("index %s out of range for %s", index, cidr)
		}
		offset = new(big.Int).Set(index)
	}

	// Base network address
	networkIP := ipNet.IP.To16()
	if networkIP == nil {
		return "", fmt.Errorf("failed to normalize IP for %s", cidr)
	}

	ipInt := new(big.Int).SetBytes(networkIP)
	ipInt.Add(ipInt, offset)

	// Convert back to IP bytes
	ipBytes := ipInt.Bytes()
	if len(ipBytes) < net.IPv6len {
		pad := make([]byte, net.IPv6len-len(ipBytes))
		ipBytes = append(pad, ipBytes...)
	}

	if isIPv4 {
		return net.IP(ipBytes[12:]).String(), nil
	}
	return net.IP(ipBytes).String(), nil
}

// countIPs returns the total number of IPs in the CIDR range.
// Works for both IPv4 and IPv6.
func CountUsableIPs(ipnet *net.IPNet) *big.Int {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	total := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(hostBits)), nil)

	if ipnet.IP.To4() == nil {
		// IPv6: all addresses usable
		return total
	}

	// IPv4 special cases
	switch ones {
	case 32:
		return big.NewInt(1)
	case 31:
		return big.NewInt(2)
	default:
		return total.Sub(total, big.NewInt(2))
	}
}
