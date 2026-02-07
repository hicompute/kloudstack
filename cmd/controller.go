/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	"github.com/hicompute/kloudstack/api/v1alpha1"
	controller "github.com/hicompute/kloudstack/pkg/controllers"
	"github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/cache"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "kloudstack kubernetes controller.",
	Long: `
	Kloudstack controller is a Kubernetes controller that manages the lifecycle of Kloudstack resources.
	It watches for changes in Kloudstack resources and performs the necessary actions to ensure that the resources are in the desired state.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ctl := controller.New()
		dc := k8s.CreateDynamicClient()
		dif := k8s.NewDynamicInformerFactory(dc, 0, "", nil)

		clusteripInformerV1alpha1 := k8s.NewDynamicInformer(dif, v1alpha1.SchemeBuilder.GroupVersion.WithResource("clusterips"))
		clusteripInformerV1alpha1.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
			AddFunc: ctl.OnAddClusterIP,
		})

		clusterippoolInformerV1alpha1 := k8s.NewDynamicInformer(dif, v1alpha1.SchemeBuilder.GroupVersion.WithResource("clusterippools"))
		clusterippoolInformerV1alpha1.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
			AddFunc: ctl.OnAddClusterIPPool,
		})

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

		go clusteripInformerV1alpha1.Run(stopCh)
		go clusterippoolInformerV1alpha1.Run(stopCh)
		go kuevirtVMInformerV1.Run(stopCh)
		go kuevirtVMIInformerV1.Run(stopCh)

		select {}
	},
}

func init() {
	rootCmd.AddCommand(controllerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// controllerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// controllerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
