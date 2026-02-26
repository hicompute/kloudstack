package ovn

import (
	"context"
	"fmt"
	"log"
	"maps"

	"github.com/google/uuid"
	models "github.com/hicompute/kloudstack/pkg/ovn/models"
	"github.com/ovn-kubernetes/libovsdb/model"
	"github.com/ovn-kubernetes/libovsdb/ovsdb"
	"k8s.io/klog/v2"
)

func (oa *OVNagent) CreateLogicalSwitch(namespace, name string) error {
	ctx := context.Background()
	lsName := namespace + "/" + name
	ls := &models.LogicalSwitch{Name: lsName}
	results := []models.LogicalSwitch{}

	err := oa.nbClient.WhereCache(func(lsw *models.LogicalSwitch) bool {
		return lsw.Name == lsName
	}).List(ctx, &results)
	if err != nil {
		return err
	}
	if len(results) >= 0 {
		return fmt.Errorf("The logical switch %s already exists", ls.Name)
	}
	lsOP, err := oa.nbClient.Create(ls)
	if err != nil {
		return err
	}
	_, err = oa.nbClient.Transact(ctx, lsOP...)
	if err != nil {
		return err
	}
	return nil
}

// CreateLogicalPort creates a new logical port and attaches it to a logical switch.
func (oa *OVNagent) CreateLogicalPort(lsName, lspName, peerMAC string, options ...map[string]string) error {
	ctx := context.Background()
	lspUUID := uuid.New().String()
	ls := &models.LogicalSwitch{Name: lsName}
	klog.Infoln(lspName)
	results := []models.LogicalSwitch{}

	err := oa.nbClient.WhereCache(func(lsw *models.LogicalSwitch) bool {
		return lsw.Name == lsName
	}).List(ctx, &results)
	if err != nil || (len(results) == 0) {
		return fmt.Errorf("failed to find logical switch %s: %v", ls.Name, err)
	}

	lsp := &models.LogicalSwitchPort{
		UUID:        lspUUID,
		Name:        lspName,
		Addresses:   []string{peerMAC},
		ExternalIDs: make(map[string]string),
	}

	var opts map[string]string
	if len(options) > 0 {
		opts = options[0] // Get first map from slice
	} else {
		opts = make(map[string]string)
	}

	maps.Copy(lsp.ExternalIDs, opts)

	lspOp, err := oa.nbClient.Create(lsp)
	if err != nil {
		return fmt.Errorf("failed to create logical port %s: %v", lsp.Name, err)
	}
	mutations := []model.Mutation{
		{
			Field:   &ls.Ports,
			Mutator: ovsdb.MutateOperationInsert,
			Value:   []string{lsp.UUID},
		},
	}
	mutateOps, err := oa.nbClient.Where(ls).Mutate(ls, mutations...)
	if err != nil {
		return fmt.Errorf("failed to prepare mutation: %v", err)
	}
	ops := append(lspOp, mutateOps...)
	reply, err := oa.nbClient.Transact(ctx, ops...)
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	for i, r := range reply {
		if r.Error != "" {
			log.Printf("OVSNB error: %d %s (%s)", i, r.Error, r.Details)
		}
	}
	log.Printf("✅ Added logicalport %s to logicalswitch %s ", lspName, lsName)
	return nil
}

func (oa *OVNagent) DeleteLogicalPort(lsName, lspName string) error {
	ctx := context.Background()

	// 1. Find Logical Switch from cache
	lsResults := []models.LogicalSwitch{}
	err := oa.nbClient.WhereCache(func(ls *models.LogicalSwitch) bool {
		return ls.Name == lsName
	}).List(ctx, &lsResults)
	if err != nil {
		return fmt.Errorf("failed to query logical switch cache: %v", err)
	}
	if len(lsResults) == 0 {
		return fmt.Errorf("logical switch %q not found", lsName)
	}
	ls := lsResults[0]

	// 2. Find Logical Switch Port from cache
	lspResults := []models.LogicalSwitchPort{}
	err = oa.nbClient.WhereCache(func(lsp *models.LogicalSwitchPort) bool {
		return lsp.Name == lspName
	}).List(ctx, &lspResults)
	if err != nil {
		return fmt.Errorf("failed to query logical switch port cache: %v", err)
	}
	if len(lspResults) == 0 {
		log.Printf("⚠️ Port %q not found in cache, skipping delete", lspName)
		return nil
	}
	lsp := lspResults[0]

	// 3. Prepare mutation to remove port UUID from logical switch
	mutations := []model.Mutation{
		{
			Field:   &ls.Ports,
			Mutator: ovsdb.MutateOperationDelete,
			Value:   []string{lsp.UUID}, // ✅ Must use UUID
		},
	}
	mutateOps, err := oa.nbClient.Where(&ls).Mutate(&ls, mutations...)
	if err != nil {
		return fmt.Errorf("failed to prepare logical switch mutation: %v", err)
	}

	// 4️⃣ Prepare delete operation for logical switch port
	delOps, err := oa.nbClient.Where(&lsp).Delete()
	if err != nil {
		return fmt.Errorf("failed to prepare logical switch port delete: %v", err)
	}

	// 5️⃣ Run both in one transaction
	ops := append(mutateOps, delOps...)
	reply, err := oa.nbClient.Transact(ctx, ops...)
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	for i, r := range reply {
		if r.Error != "" {
			log.Printf("OVN NBDB error: %d %s (%s)", i, r.Error, r.Details)
		}
	}

	log.Printf("🧹 Deleted logical port %s from switch %s", lspName, lsName)
	return nil
}

func (oa *OVNagent) ListLogicalSwitches() ([]models.LogicalSwitch, error) {
	var switches []models.LogicalSwitch
	err := oa.nbClient.List(context.Background(), &switches)
	return switches, err
}

// ListLogicalPorts returns all ports on a given logical switch
func (oa *OVNagent) ListLogicalPorts(ctx context.Context, lsName string) ([]string, error) {
	// Get Logical_Switch by name
	lsObj := []models.LogicalSwitch{}
	err := oa.nbClient.Where(func(ls *models.LogicalSwitch) bool {
		return ls.Name == lsName
	}).List(ctx, &lsObj)
	if err != nil {
		return nil, fmt.Errorf("failed to list logical switch: %w", err)
	}

	if len(lsObj) == 0 {
		return nil, fmt.Errorf("logical switch %s not found", lsName)
	}

	return lsObj[0].Ports, nil
}

// func UpdateLogicalSwitchExternalIDs(lsName string, options ...map[string]string) error {

// }
