package ovs

import (
	"context"
	"log"
	"strings"

	helper "github.com/hicompute/kloudstack/pkg/helpers"
	"github.com/hicompute/kloudstack/pkg/metrics"
	ovsModels "github.com/hicompute/kloudstack/pkg/ovs/models"
	"github.com/ovn-kubernetes/libovsdb/client"
	"github.com/ovn-kubernetes/libovsdb/model"
)

type OvsAgent struct {
	ovsClient client.Client
}

func CreateOVSagent() (*OvsAgent, error) {
	dbModel, err := model.NewClientDBModel("Open_vSwitch", map[string]model.Model{
		ovsModels.BridgeTable:    &ovsModels.Bridge{},
		ovsModels.PortTable:      &ovsModels.Port{},
		ovsModels.InterfaceTable: &ovsModels.Interface{},
	})
	dbModel.SetIndexes(map[string][]model.ClientIndex{
		ovsModels.PortTable: {{Columns: []model.ColumnKey{
			{Column: "external_ids", Key: "iface-id"},
		}}},
	})
	if err != nil {
		log.Printf("failed to create DB model: %v", err)
		return nil, err
	}

	ovsClient, err := client.NewOVSDBClient(
		dbModel,
		client.WithEndpoint("unix:/var/run/openvswitch/db.sock"),
	)
	if err != nil {
		log.Printf("failed to create OVS client: %v", err)
		return nil, err
	}

	ctx := context.Background()
	if err := ovsClient.Connect(ctx); err != nil {
		log.Printf("failed to connect to OVSDB: %v", err)
		return nil, err
	}
	_, err = ovsClient.MonitorAll(ctx)
	if err != nil {
		log.Printf("failed to monitor OVSDB: %v", err)
		return nil, err
	}
	return &OvsAgent{
		ovsClient: ovsClient,
	}, nil
}

func (oa *OvsAgent) Close() {
	oa.ovsClient.Disconnect()
}

func (oa *OvsAgent) Start(ctx context.Context) {
	oa.ovsClient.Cache().AddEventHandler(&ovsEventHandler{agent: oa})
}

func (oa *OvsAgent) updateInterfaceStats(iface *ovsModels.Interface) {
	if iface.Statistics == nil {
		return
	}

	stats := make(map[string]int64)
	for k, v := range iface.Statistics {
		stats[k] = int64(v)
	}

	if iface.ExternalIDs == nil {
		return
	}

	ifaceID := iface.ExternalIDs["iface-id"]

	parts := strings.Split(ifaceID, "_")

	if len(parts) != 3 {
		return
	}

	labels := map[string]string{
		"vm_namespace": parts[0],
		"vm":           helper.ExtractVMName(parts[1]),
		"iface":        parts[2],
	}

	if v, ok := stats["rx_bytes"]; ok {
		metrics.RxBytes.With(labels).Set(float64(v))
	}
	if v, ok := stats["tx_bytes"]; ok {
		metrics.TxBytes.With(labels).Set(float64(v))
	}
}
