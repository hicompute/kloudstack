/*
Copyright © 2026 hicompute <kloudstack@hicompute.ir>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hicompute/kloudstack/pkg/metrics"
	"github.com/hicompute/kloudstack/pkg/ovs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var metricExporterCmd = &cobra.Command{
	Use:   "metric-exporter",
	Short: "Start Metric Exporter",
	Long:  `Start a Metric Exporter.`,

	Run: func(cmd *cobra.Command, args []string) {
		metrics.RegisterOVSMetrics()

		agent, err := ovs.CreateOVSagent()
		if err != nil {
			panic(err)
		}
		ctx := context.Background()
		agent.Start(ctx)

		http.Handle("/metrics", promhttp.Handler())
		port := viper.GetInt32("port")
		http.ListenAndServe(":"+fmt.Sprintf("%d", port), nil)
	},
}

func init() {
	metricExporterCmd.Flags().Int32P("port", "p", 9476, "Exporter http port.")

	viper.BindPFlag("PORT", metricExporterCmd.Flags().Lookup("port"))

	rootCmd.AddCommand(metricExporterCmd)
}
