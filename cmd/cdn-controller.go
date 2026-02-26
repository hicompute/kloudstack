/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	"github.com/hicompute/kloudstack/api/network/v1alpha1"
	controller "github.com/hicompute/kloudstack/pkg/controllers/cdn"
	"github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/cache"
)

// controllerCmd represents the controller command
var cdnControllerCmd = &cobra.Command{
	Use:   "cdncontroller",
	Short: "kloudstack CDN kubernetes controller.",
	Long: `
	Kloudstack cdn controller is a Kubernetes controller that manages the lifecycle of Kloudstack cdn resources.
	It watches for changes in Kloudstack resources and performs the necessary actions to ensure that the resources are in the desired state.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ctl := controller.New()
		dc := k8s.CreateDynamicClient()
		dif := k8s.NewDynamicInformerFactory(dc, 0, "", nil)

		gatewayInformerV1alpha1 := k8s.NewDynamicInformer(dif, v1alpha1.SchemeBuilder.GroupVersion.WithResource("gateways"))
		gatewayInformerV1alpha1.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
			AddFunc: ctl.OnAddGateway,
		})

		// httpRouteInformerV1alpha1 := k8s.NewDynamicInformer(dif, v1alpha1.SchemeBuilder.GroupVersion.WithResource("httproutes"))
		// httpRouteInformerV1alpha1.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		// 	AddFunc: ctl.On,
		// })

		stopCh := make(chan struct{})

		defer close(stopCh)

		go gatewayInformerV1alpha1.Run(stopCh)

		select {}
	},
}

func init() {
	cdnControllerCmd.Flags().StringP("redis_host", "rh", "localhost", "redis host address.")
	cdnControllerCmd.Flags().Int32P("redis_port", "rp", 6379, "redis port.")
	cdnControllerCmd.Flags().IntP("redis_db", "rdb", 0, "redis database.")
	cdnControllerCmd.Flags().StringP("redis_password", "rpwd", "", "redis password.")
	cdnControllerCmd.Flags().StringP("hostname", "h", "tlscdn.ir", "cdn hostname.")
	cdnControllerCmd.Flags().StringP("certmanager-clusterissuer", "i", "kloudstack-letsencrypt-webhook", "certmanager clusterissuer name.")

	viper.BindPFlag("REDIS_HOST", cdnControllerCmd.Flags().Lookup("redis_host"))
	viper.BindPFlag("REDIS_PORT", cdnControllerCmd.Flags().Lookup("redis_port"))
	viper.BindPFlag("REDIS_DB", cdnControllerCmd.Flags().Lookup("redis_db"))
	viper.BindPFlag("REDIS_PASSWORD", cdnControllerCmd.Flags().Lookup("redis_password"))
	viper.BindPFlag("CDN_HOSTNAME", cdnControllerCmd.Flags().Lookup("hostname"))
	viper.BindPFlag("CERTMANAGER_CLUSTER_ISSUER", cdnControllerCmd.Flags().Lookup("certmanager-clusterissuer"))

	rootCmd.AddCommand(cdnControllerCmd)
}
