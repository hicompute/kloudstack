/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	"fmt"
	"net"
	"os"

	dhcpd "github.com/hicompute/kloudstack/pkg/daemons/dhcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dhcpdCmd = &cobra.Command{
	Use:   "dhcpd",
	Short: "Start DHCP server",
	Long: `Start a DHCP server to provide automatic IP address configuration to clients using IPAM.

Examples:
  Start DHCP server on default interface:
    kloudstack dhcpd

  Start DHCP server on specific interface with custom DNS:
    kloudstack dhcpd -I eth0 -d 1.1.1.1,1.0.0.1

  Start IPv6 DHCP server:
    kloudstack dhcpd --family v6`,
	PreRun: func(cmd *cobra.Command, args []string) {
		ipFamily := viper.GetString("IPFAMILY")
		if ipFamily != "v4" && ipFamily != "v6" {
			fmt.Printf("Error: family must be 'v4' or 'v6', got '%s'\n", ipFamily)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		server := viper.GetString("SERVER")
		if net.ParseIP(server) == nil {
			fmt.Printf("Error: server must be a valid IP address, got '%s'\n", server)
			os.Exit(1)
		}
		ipFamily := viper.GetString("IPFAMILY")
		networkInterface := viper.GetString("INTERFACE")

		if ipFamily == "v4" {
			if err := dhcpd.StartV4Server(networkInterface); err != nil {
				fmt.Printf("Error on starting dhcp daemon: %v", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("IP %s dhcp server would release in future versions.", ipFamily)
		}
	},
}

func init() {
	dhcpdCmd.Flags().StringP("interface", "i", "br-ext", "network interface name.")
	dhcpdCmd.Flags().StringP("listen", "l", "0.0.0.0", "Listen on this IP address.")
	dhcpdCmd.Flags().StringP("family", "f", "v4", "ip family (v4 or v6).")
	dhcpdCmd.Flags().StringP("mac_prefix", "m", "", "only effects on mac addresses with specific prefix.")
	dhcpdCmd.Flags().StringP("dns", "d", "8.8.8.8,8.8.4.4", "comma separated dns servers list.")
	dhcpdCmd.Flags().StringP("server", "s", "", "dhcp server address (required).")

	dhcpdCmd.MarkFlagRequired("server")

	viper.BindPFlag("INTERFACE", dhcpdCmd.Flags().Lookup("interface"))
	viper.BindPFlag("LISTENADDRESS", dhcpdCmd.Flags().Lookup("listen"))
	viper.BindPFlag("IPFAMILY", dhcpdCmd.Flags().Lookup("family"))
	viper.BindPFlag("MACPREFIX", dhcpdCmd.Flags().Lookup("mac_prefix"))
	viper.BindPFlag("DNS", dhcpdCmd.Flags().Lookup("dns"))
	viper.BindPFlag("SERVER", dhcpdCmd.Flags().Lookup("server"))

	rootCmd.AddCommand(dhcpdCmd)
}
