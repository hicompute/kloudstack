package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ovs_port_rx_bytes",
			Help: "Received bytes on OVS port",
		},
		[]string{"vm_namespace", "vm", "iface"},
	)
	TxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ovs_port_tx_bytes",
			Help: "Transmitted bytes on OVS port",
		},
		[]string{"vm_namespace", "vm", "iface"},
	)
)

func RegisterOVSMetrics() {
	prometheus.MustRegister(RxBytes, TxBytes)
}
