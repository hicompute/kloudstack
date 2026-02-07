package ovn

import (
	"context"

	models "github.com/hicompute/kloudstack/pkg/ovn/models"
	"github.com/ovn-kubernetes/libovsdb/client"
	"github.com/ovn-kubernetes/libovsdb/model"
)

type OVNagent struct {
	nbClient client.Client
}

func CreateOVNagent(nbEndpoint string) (*OVNagent, error) {
	// Define database model
	dbModel, err := model.NewClientDBModel("OVN_Northbound", map[string]model.Model{
		models.LogicalSwitchTable:     &models.LogicalSwitch{},
		models.LogicalSwitchPortTable: &models.LogicalSwitchPort{},
		// Add other table mappings
	})
	if err != nil {
		return nil, err
	}
	dbModel.SetIndexes(map[string][]model.ClientIndex{
		models.LogicalSwitchTable: {
			{Columns: []model.ColumnKey{{Column: "name"}}},
		},
	})

	// Create client with connection options
	nbClient, err := client.NewOVSDBClient(dbModel, client.WithEndpoint(nbEndpoint))
	if err != nil {
		return nil, err
	}

	// Establish connection
	ctx := context.Background()
	err = nbClient.Connect(ctx)
	if err != nil {
		return nil, err
	}

	// Start monitoring for cache updates
	_, err = nbClient.MonitorAll(ctx)
	if err != nil {
		return nil, err
	}

	return &OVNagent{nbClient: nbClient}, nil
}

func (o *OVNagent) Close() {
	o.nbClient.Disconnect()
}
