package ovs

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	ovsModel "github.com/hicompute/kloudstack/pkg/ovs/models"
	"github.com/ovn-kubernetes/libovsdb/model"
	"github.com/ovn-kubernetes/libovsdb/ovsdb"
	"k8s.io/klog/v2"
)

func (oa *OvsAgent) AddPort(bridgeName, portName, ifaceType, ifaceId string) error {
	ctx := context.Background()
	ifaceUUID := uuid.New().String()
	portUUID := uuid.New().String()
	bridge := &ovsModel.Bridge{Name: bridgeName}
	if err := oa.ovsClient.Get(ctx, bridge); err != nil {
		return fmt.Errorf("failed to get bridge %q: %v", bridgeName, err)
	}

	iface := &ovsModel.Interface{
		UUID: ifaceUUID,
		Name: portName,
		Type: ifaceType,
		// MAC:  &macAddress,
		ExternalIDs: map[string]string{
			"iface-id": ifaceId,
		},
	}

	ifaceOp, err := oa.ovsClient.Create(iface)
	if err != nil {
		return fmt.Errorf("failed to create interface: %v", err)
	}
	externalIds := map[string]string{
		"iface-id": ifaceId,
	}
	port := &ovsModel.Port{
		UUID:        portUUID,
		Name:        portName,
		Interfaces:  []string{iface.UUID},
		ExternalIDs: externalIds,
	}
	portOp, err := oa.ovsClient.Create(port)
	if err != nil {
		return fmt.Errorf("failed to create port: %v", err)
	}

	mutations := []model.Mutation{
		{
			Field:   &bridge.Ports,
			Mutator: ovsdb.MutateOperationInsert,
			Value:   []string{port.UUID},
		},
	}
	mutateOps, err := oa.ovsClient.Where(bridge).Mutate(bridge, mutations...)
	if err != nil {
		return fmt.Errorf("failed to prepare mutation: %v", err)
	}
	ops := append(ifaceOp, append(portOp, mutateOps...)...)
	reply, err := oa.ovsClient.Transact(ctx, ops...)
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	for i, r := range reply {
		if r.Error != "" {
			klog.Infof("OVSDB error: %d %s (%s)", i, r.Error, r.Details)
		}
	}
	klog.Infof("✅ Added port %s to bridge %s (type=%s)", portName, bridgeName, ifaceType)
	return nil
}

func (oa *OvsAgent) DelPort(bridgeName, ifaceId string) error {
	ctx := context.Background()

	externalIds := map[string]string{
		"iface-id": ifaceId,
	}
	// 1. Get port and bridge objects
	bridge := &ovsModel.Bridge{Name: bridgeName}
	if err := oa.ovsClient.Get(ctx, bridge); err != nil {
		return fmt.Errorf("failed to find bridge %s: %v", bridgeName, err)
	}
	ports := []ovsModel.Port{}
	if err := oa.ovsClient.Where(&ovsModel.Port{ExternalIDs: externalIds}).List(ctx, &ports); err != nil {
		return fmt.Errorf("failed to find port with iface-id %s: %v", ifaceId, err)
	}

	if len(ports) == 0 {
		return fmt.Errorf("port not found with iface-id %s", ifaceId)
	}
	port := ports[0]
	// 2. Mutate the bridge to remove the port UUID from its Ports set
	mutations := []model.Mutation{
		{
			Field:   &bridge.Ports,
			Mutator: ovsdb.MutateOperationDelete,
			Value:   []string{port.UUID},
		},
	}

	bridgeOps, err := oa.ovsClient.Where(bridge).Mutate(bridge, mutations...)
	if err != nil {
		return fmt.Errorf("failed to prepare bridge mutation: %v", err)
	}
	// 3. Delete the port itself (and the interface)
	portOp, err := oa.ovsClient.Where(&port).Delete()
	if err != nil {
		return fmt.Errorf("failed to prepare port delete: %v", err)
	}
	// 4. Also delete the Interface row(s) belonging to the port
	for _, ifaceUUID := range port.Interfaces {
		iface := &ovsModel.Interface{UUID: ifaceUUID}
		ifaceOp, err := oa.ovsClient.Where(iface).Delete()
		if err != nil {
			return fmt.Errorf("failed to prepare interface delete: %v", err)
		}
		portOp = append(portOp, ifaceOp...)
	}

	// 5. Run all operations in one transaction
	ops := append(bridgeOps, portOp...)
	reply, err := oa.ovsClient.Transact(ctx, ops...)
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	for i, r := range reply {
		if r.Error != "" {
			klog.Infof("OVSDB error: %d %s (%s), rows: %v", i, r.Error, r.Details, r.Rows)
		}
	}

	klog.Infof("🧹 Deleted port %v from bridge %s", externalIds, bridgeName)
	return nil
}
