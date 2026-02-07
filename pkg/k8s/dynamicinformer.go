package k8s

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

func NewDynamicInformerFactory(client *dynamic.DynamicClient, resyncPeriod time.Duration, namespace string, tweakListOptions dynamicinformer.TweakListOptionsFunc) dynamicinformer.DynamicSharedInformerFactory {
	return dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, resyncPeriod, namespace, tweakListOptions)
}

func NewDynamicInformer(dif dynamicinformer.DynamicSharedInformerFactory, gvr schema.GroupVersionResource) cache.SharedIndexInformer {
	return dif.ForResource(gvr).Informer()
}
