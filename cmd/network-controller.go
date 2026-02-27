/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	"github.com/hicompute/kloudstack/api/network/v1alpha1"
	controller "github.com/hicompute/kloudstack/pkg/controllers/network"
	"github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/cache"
)

// networkControllerCmd represents the network controller command
var networkControllerCmd = &cobra.Command{
	Use:   "network-controller",
	Short: "kloudstack network kubernetes controller.",
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

		logicalSwitchInformerV1Alpha1 := k8s.NewDynamicInformer(dif, v1alpha1.SchemeBuilder.GroupVersion.WithResource("logicalswitches"))
		logicalSwitchInformerV1Alpha1.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
			AddFunc: ctl.OnAddLogicalSwitch,
		})

		stopCh := make(chan struct{})

		defer close(stopCh)

		go clusteripInformerV1alpha1.Run(stopCh)
		go clusterippoolInformerV1alpha1.Run(stopCh)

		go logicalSwitchInformerV1Alpha1.Run(stopCh)

		select {}
	},
}

func init() {
	rootCmd.AddCommand(networkControllerCmd)
}
