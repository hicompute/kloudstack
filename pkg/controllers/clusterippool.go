package controller

import (
	"context"
	"fmt"
	"net"
	"reflect"

	"github.com/hicompute/kloudstack/api/v1alpha1"
	netutils "github.com/hicompute/kloudstack/pkg/net_utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (c *Controller) OnAddClusterIPPool(req any, isInInitialList bool) {
	if isInInitialList {
		return
	}
	ctx := context.Background()
	obj := req.(*unstructured.Unstructured)

	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	var pool v1alpha1.ClusterIPPool
	if err := c.Client.Get(ctx, key, &pool); err != nil {
		klog.Errorf("%v", err)
		return
	}

	_, ipnet, err := net.ParseCIDR(pool.Spec.CIDR)
	if err != nil {
		// Set a degraded condition
		meta.SetStatusCondition(&pool.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "InvalidCIDR",
			Message: fmt.Sprintf("Invalid CIDR: %v", err),
		})
		_ = c.Client.Status().Update(ctx, &pool)
		return
	}
	totalIPs := netutils.CountUsableIPs(ipnet)
	newStatus := pool.Status.DeepCopy()
	newStatus.TotalIPs = totalIPs.String()
	newStatus.FreeIPs = totalIPs.String()
	if reflect.DeepEqual(&pool.Status, newStatus) {
		return // no changes
	}
	pool.Status = *newStatus
	if err := c.Client.Status().Update(ctx, &pool); err != nil {
		klog.Errorf("Error on updating cluster ip pool status: %v", err)
	}
}
