/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	ovncnid "github.com/hicompute/kloudstack/pkg/daemons/ovn-cni-server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

var ovncnidCmd = &cobra.Command{
	Use:   "ovncnid",
	Short: "Start OVS/OVN cni daemon.",
	Long:  `Start OVS/OVN cni daemon. This command starts the OVS/OVN cni daemon. It is used to start the OVS/OVN cni daemon. example: ovncnid -f /var/run/kloudstack/ovn-cni.sock`,
	Run: func(cmd *cobra.Command, args []string) {
		cniSocketFile := viper.GetString("CNI_SOCKET_FILE")
		if err := ovncnid.Start(cniSocketFile); err != nil {
			klog.Fatalf("Error on starting ovn cni daemon: %v", err)
		}
	},
}

func init() {
	ovncnidCmd.Flags().StringP("cni-socket-file", "f", "/var/run/kloudstack/ovn-cni.sock", "cni socket file path.")
	viper.BindPFlag("CNI_SOCKET_FILE", ovncnidCmd.Flags().Lookup("cni-socket-file"))
	rootCmd.AddCommand(ovncnidCmd)
}
