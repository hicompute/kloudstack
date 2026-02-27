/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	controller "github.com/hicompute/kloudstack/pkg/controllers/vm"
	"github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/cache"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// computeControllerCmd represents the vm controller command
var vmControllerCmd = &cobra.Command{
	Use:   "kubevirt-controller",
	Short: "kloudstack kubevirt virthal machine controller.",
	Long: `
	Kloudstack kubevirt controller is a Kubernetes controller that manages the lifecycle of kubevirt resources.
	It watches for changes in kubevirt resources and performs the necessary kloudstack related actions to ensure that the resources are in the desired state.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ctl := controller.New()
		dc := k8s.CreateDynamicClient()
		dif := k8s.NewDynamicInformerFactory(dc, 0, "", nil)

		kuevirtVMInformerV1 := k8s.NewDynamicInformer(dif, kubevirtv1.GroupVersion.WithResource("virtualmachines"))
		kuevirtVMInformerV1.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
			AddFunc:    ctl.OnAddKubevirtVM,
			DeleteFunc: ctl.OnDeleteKubevirtVM,
		})

		kuevirtVMIInformerV1 := k8s.NewDynamicInformer(dif, kubevirtv1.GroupVersion.WithResource("virtualmachineinstances"))
		kuevirtVMIInformerV1.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
			UpdateFunc: ctl.OnUpdateVMI,
		})

		stopCh := make(chan struct{})

		defer close(stopCh)

		go kuevirtVMInformerV1.Run(stopCh)
		go kuevirtVMIInformerV1.Run(stopCh)

		select {}
	},
}

func init() {
	rootCmd.AddCommand(vmControllerCmd)
}
