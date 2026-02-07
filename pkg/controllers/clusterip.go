package controller

import (
	"github.com/hicompute/kloudstack/api/v1alpha1"
	"k8s.io/klog/v2"
)

func (c *Controller) OnAddClusterIP(obj any, isInInitialList bool) {
	clusterIP, ok := obj.(*v1alpha1.ClusterIP)
	if !ok {
		klog.Errorf("not correct")
	}
	klog.Infof("%v", clusterIP)
}
