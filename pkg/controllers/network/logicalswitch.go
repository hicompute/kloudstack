package controller

import (
	"context"

	"github.com/hicompute/kloudstack/api/network/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (c *Controller) OnAddLogicalSwitch(req any, isInInitialList bool) {
	ctx := context.Background()
	obj := req.(*unstructured.Unstructured)
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	var ls v1alpha1.LogicalSwitch

	if err := c.Client.Get(ctx, key, &ls); err != nil {
		klog.Errorf("%v", err)
		return
	}
	if err := c.ovnAgent.CreateLogicalSwitch(obj.GetNamespace(), obj.GetName()); err != nil {
		klog.Errorf("Failed to create logical switch: %v", err)
	}
}
