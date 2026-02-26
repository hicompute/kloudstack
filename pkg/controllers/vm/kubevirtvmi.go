package controller

import (
	"context"

	"github.com/hicompute/kloudstack/api/network/v1alpha1"
	"github.com/samber/lo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Controller) OnUpdateVMI(_, req any) {
	ctx := context.Background()
	obj := req.(*unstructured.Unstructured)
	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	var vmi kubevirtv1.VirtualMachineInstance

	if err := c.Client.Get(ctx, key, &vmi); err != nil {
		klog.Errorf("%v", err)
		return
	}

	if vmi.Status.Phase != kubevirtv1.Running {
		return
	}

	interfaces := vmi.Status.Interfaces
	if len(interfaces) == 0 {
		return
	}

	clusterIPList := v1alpha1.ClusterIPList{}
	if err := c.Client.List(ctx, &clusterIPList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.resource", vmi.Namespace+"/"+vmi.Name),
		Limit:         -1,
	}); err != nil {
		klog.Errorf("Error on getting cip list: %v", err)
	}

	releaseIpList := lo.Reduce(clusterIPList.Items, func(result []v1alpha1.ClusterIP, item v1alpha1.ClusterIP, _ int) []v1alpha1.ClusterIP {
		_, ok := lo.Find(interfaces, func(i kubevirtv1.VirtualMachineInstanceNetworkInterface) bool {
			return i.MAC == item.Spec.Mac
		})
		if !ok {
			historyItem := v1alpha1.ClusterIPHistory{
				Mac:        item.Spec.Mac,
				ReleasedAt: v1.Now(),
				Interface:  item.Spec.Interface,
				Resource:   item.Spec.Resource,
			}
			if len(item.Status.History) <= 0 {
				item.Status.History = []v1alpha1.ClusterIPHistory{historyItem}
			} else {
				item.Status.History = append(item.Status.History, historyItem)
			}
			item.Spec.Mac = ""
			item.Spec.Interface = ""
			result = append(result, item)
		}
		return result
	}, []v1alpha1.ClusterIP{})

	for i := range releaseIpList {
		if err := c.Client.Update(ctx, &releaseIpList[i]); err != nil {
			klog.Errorf("Error on updating cip: %v", err)
		}
	}
}
