/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	"fmt"

	"github.com/hicompute/kloudstack/pkg/webhooks/gatekeeper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gatekeeperWebhookCmd = &cobra.Command{
	Use:   "gatekeeper",
	Short: "Start gatekeeper authorization webhook.",
	Long:  `Start gatekeeper authorization webhook. This command starts the gatekeeper webhook server which is responsible for authorizing incoming requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		webook := gatekeeper.New()
		if err := webook.Start(); err != nil {
			fmt.Printf("error on starting gatekeeper webhook: %v", err)
		}
	},
}

func init() {
	gatekeeperWebhookCmd.Flags().Int32P("port", "p", 443, "listening port.")
	gatekeeperWebhookCmd.Flags().BoolP("tls", "t", true, "tls mode.")
	gatekeeperWebhookCmd.Flags().StringP("tls-cert-file", "c", "/tls/tls.crt", "tls cert file path.")
	gatekeeperWebhookCmd.Flags().StringP("tls-key-file", "k", "/tls/tls.key", "tls key file path.")

	viper.BindPFlag("PORT", gatekeeperWebhookCmd.Flags().Lookup("port"))
	viper.BindPFlag("TLS", gatekeeperWebhookCmd.Flags().Lookup("tls"))
	viper.BindPFlag("TLS_CERT_FILE", gatekeeperWebhookCmd.Flags().Lookup("tls-cert-file"))
	viper.BindPFlag("TLS_KEY_FILE", gatekeeperWebhookCmd.Flags().Lookup("tls-key-file"))

	rootCmd.AddCommand(gatekeeperWebhookCmd)
}
