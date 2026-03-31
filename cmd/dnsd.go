package cmd

import (
	"github.com/hicompute/kloudstack/pkg/daemons/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

var dnsdCmd = &cobra.Command{
	Use:   "dnsd",
	Short: "Start DHCP GRPC server",
	Long:  `Start the Kloudstack DNS server daemon (coredns grpc plugin).`,
	Run: func(cmd *cobra.Command, args []string) {
		socketFile := viper.GetString("SOCKET_FILE")
		if err := dns.Start(socketFile); err != nil {
			klog.Fatalf("Error on starting dns grpc daemon: %v", err)
		}
	},
}

func init() {
	dnsdCmd.Flags().StringP("socket-file", "f", "/var/run/kloudstack/dns.sock", "socket file path.")
	dnsdCmd.Flags().StringP("scylladb-keyspace", "k", "kloudstack_dns", "scylladb keyspace.")
	dnsdCmd.Flags().StringSliceP("scylladb-hosts", "H", []string{"127.0.0.1"}, "ScyllaDB cluster hosts")
	dnsdCmd.Flags().StringP("scylladb-datacenter", "d", "datacenter1", "scylladb datacenter.")
	dnsdCmd.Flags().Uint32P("scylladb-port", "p", 9042, "ScyllaDB port")
	dnsdCmd.Flags().StringP("dns-mname", "m", "ns1.ns53.ir.", "nameserver m name.")
	dnsdCmd.Flags().StringP("dns-rname", "r", "admin.ns53.ir.", "nameserver r name.")

	viper.BindPFlag("SOCKET_FILE", dnsdCmd.Flags().Lookup("socket-file"))
	viper.BindPFlag("SCYLLADB_HOSTS", dnsdCmd.Flags().Lookup("scylladb-hosts"))
	viper.BindPFlag("SCYLLADB_DATACENTER", dnsdCmd.Flags().Lookup("scylladb-datacenter"))
	viper.BindPFlag("SCYLLADB_KEYSPACE", dnsdCmd.Flags().Lookup("scylladb-keyspace"))
	viper.BindPFlag("SCYLLADB_PORT", dnsdCmd.Flags().Lookup("scylladb-port"))
	viper.BindPFlag("DNS_M_NAME", dnsdCmd.Flags().Lookup("dns-mname"))
	viper.BindPFlag("DNS_R_NAME", dnsdCmd.Flags().Lookup("dns-rname"))
	rootCmd.AddCommand(dnsdCmd)
}
