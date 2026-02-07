package ovs

import (
	ovsModels "github.com/hicompute/kloudstack/pkg/ovs/models"
	"github.com/ovn-kubernetes/libovsdb/model"
)

type ovsEventHandler struct {
	agent *OvsAgent
}

func (h *ovsEventHandler) OnAdd(table string, m model.Model) { h.handle(table, m) }
func (h *ovsEventHandler) OnUpdate(table string, old, new model.Model) {
	h.handle(table, new)
}
func (h *ovsEventHandler) OnDelete(table string, m model.Model) {}

func (h *ovsEventHandler) handle(tabel string, m model.Model) {
	switch obj := m.(type) {

	case *ovsModels.Interface:
		h.agent.updateInterfaceStats(obj)
	}
}
