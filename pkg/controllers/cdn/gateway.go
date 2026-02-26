package controller

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	cmapi "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	v1alpha1 "github.com/hicompute/kloudstack/api/cdn/v1alpha1"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (c *Controller) OnAddGateway(req any, isInitial bool) {
	ctx := context.Background()
	obj := req.(*unstructured.Unstructured)

	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	var gateway v1alpha1.Gateway
	if err := c.Client.Get(ctx, key, &gateway); err != nil {
		klog.Errorf("%v", err)
		return
	}

	data := map[string]any{
		"name":        gateway.GetName(),
		"namespace":   gateway.GetNamespace(),
		"UID":         gateway.GetUID(),
		"upstreams":   gateway.Spec.Upstreams,
		"waf_enabled": gateway.Spec.WafEnabled,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		klog.Errorf("Failed to marshal data: %v", err)
	}

	redisKey := gateway.Spec.Domain
	if redisKey == "" {
		redisKey = strings.ReplaceAll(string(gateway.GetUID()), "-", "") + "." + viper.GetString("CDN_HOSTNAME")
	}

	err = c.RedisClient.Set(context.Background(), redisKey, jsonData, 0).Err()
	if err != nil {
		klog.Errorf("Error on Redis set for gateway %s/%s: %v", gateway.GetNamespace(), gateway.GetName(), err)
	}

	if gateway.Spec.Domain != "" && gateway.Spec.Tls == "auto" {
		err = c.createLetsEncryptWildCardCertificate(gateway.Namespace, gateway.Name, gateway.Spec.Domain)
		if err != nil {
			klog.Errorf("failed to create certificate for cdnGateway %s/%s: %v", gateway.Namespace, gateway.Name, err)
		}
		// update gateway status with certificate name or data
	}

}

func (c *Controller) OnUpdateGateway(prev any, req any) {
	prevSpec := getSpec(prev.(*unstructured.Unstructured))
	if reflect.DeepEqual(
		prevSpec,
		getSpec(req.(*unstructured.Unstructured)),
	) {
		return
	}

	ctx := context.Background()
	obj := req.(*unstructured.Unstructured)

	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	var gateway v1alpha1.Gateway
	if err := c.Client.Get(ctx, key, &gateway); err != nil {
		klog.Errorf("%v", err)
		return
	}

	// find all httproutes associated to the gateway.
	httpRoutes, err := c.findHttpRoutesByGatewayName(gateway.GetName(), gateway.GetNamespace())
	if err != nil {
		klog.Errorf("Error finding httproutes by gateway name: %v", err)
		return
	}
	if len(httpRoutes.Items) == 0 {
		klog.Info("No associated httproutes found for gateway")
		return
	}

	for _, httpRoute := range httpRoutes.Items {
		deletedUpstream := true
		for _, upstreams := range gateway.Spec.Upstreams {
			if httpRoute.Spec.UpstreamName == upstreams.Name {
				deletedUpstream = false
			}
		}
		if deletedUpstream {
			httpRoute.Spec.UpstreamName = ""
		}

		annotations := httpRoute.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["gateway.cdn.kloudstack.ir/updated"] = time.Now().Format(time.RFC3339Nano)

		uo := &v1alpha1.HTTPRoute{
			Spec: httpRoute.Spec,
		}
		uo.SetName(httpRoute.Name)
		uo.SetNamespace(httpRoute.Namespace)
		uo.SetAnnotations(annotations)
		uo.SetResourceVersion(httpRoute.GetResourceVersion())
		if err = c.Client.Update(ctx, uo); err != nil {
			klog.Errorf("Error on httproute apply changes: %v", err)
			return
		}
		klog.Infof("Updated httproute %s for gateway %s", httpRoute.GetName(), gateway.GetName())
	}

	prevDomain, _, _ := unstructured.NestedString(prev.(*unstructured.Unstructured).Object, "spec", "domain")
	if gateway.Spec.Domain != "" && gateway.Spec.Tls == "auto" && gateway.Spec.Domain != prevDomain {
		if prevDomain != "" {
			// delete certificate for old domain
			if err = c.deleteLetsEncryptWildCardCertificate(gateway.Namespace, gateway.Name); err != nil {
				klog.Errorf("failed to delete old certificate for gateway %s/%s : %v", gateway.Name, gateway.Namespace, err)
			}
			// invalidate cert cache for old domain.

			err = c.RedisClient.Publish(context.Background(), "invalidate_gateway_cache", prevDomain).Err()
			if err != nil {
				klog.Errorf("[Update gateway] certiticate cache invalidation, publish message for gateway %s/%s was unsuccessful: %v", gateway.GetName(), gateway.GetNamespace(), err)
			}
		}
		err = c.createLetsEncryptWildCardCertificate(gateway.Namespace, gateway.Name, gateway.Spec.Domain)
		if err != nil {
			klog.Errorf("failed to create certificate for cdnGateway %s/%s: %v", gateway.Namespace, gateway.Name, err)
		}
		// update gateway status with certificate name or data
	}
}

func (c *Controller) OnDeleteGateway(obj any) {
	ctx := context.Background()
	u := obj.(*unstructured.Unstructured)
	name := u.GetName()
	namespace := u.GetNamespace()
	if err := c.Client.DeleteAllOf(ctx, &v1alpha1.HTTPRoute{}, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     namespace,
			FieldSelector: fields.OneTermEqualSelector("spec.gateway.name", u.GetName()),
		},
	}); err != nil {
		klog.Errorf("Error on deleting associated httroutes for gateway %s on namespace %s: %v", name, namespace, err)
	}

	redisKey := strings.ReplaceAll(string(u.GetUID()), "-", "") + "." + viper.GetString("CDN_HOSTNAME")
	if err := c.RedisClient.Del(context.Background(), redisKey).Err(); err != nil {
		klog.Errorf("Error on deleting associated redis key for gateway %s on namespace %s: %v", name, namespace, err)
	}
	if err := c.RedisClient.Publish(context.Background(), "invalidate_gateway_cache", redisKey).Err(); err != nil {
		klog.Errorf("[Delete gateway] cache invalidation, publish message for gateway %s in namespace %s was unsuccessful: %v", name, namespace, err)
		return
	}
}

func (c *Controller) createLetsEncryptWildCardCertificate(namespace, name, hostname string) error {
	cert := &cmapi.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cmapi.CertificateSpec{
			SecretName: name + "-tls-cert",
			IssuerRef: cmmeta.ObjectReference{
				Name: viper.GetString("CERTMANAGER_CLUSTER_ISSUER"),
				Kind: "ClusterIssuer",
			},
			DNSNames: []string{hostname, "*." + hostname},
			SecretTemplate: &cmapi.CertificateSecretTemplate{
				Annotations: map[string]string{
					"cdn.kloudstack.ir/hostname": hostname,
				},
			},
		},
	}

	return c.Client.Create(context.Background(), cert)
}

func (c *Controller) deleteLetsEncryptWildCardCertificate(namespace string, name string) error {
	return c.Client.Delete(context.Background(), &cmapi.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	})
}

func (c *Controller) findHttpRoutesByGatewayName(name string, namespace string) (*v1alpha1.HTTPRouteList, error) {
	ctx := context.Background()
	var httpRouteList v1alpha1.HTTPRouteList
	if err := c.Client.List(ctx, &httpRouteList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.gateway.name", name),
	}); err != nil {
		return nil, err
	}
	return &httpRouteList, nil
}

func getSpec(obj *unstructured.Unstructured) interface{} {
	spec, found, err := unstructured.NestedFieldCopy(obj.Object, "spec")
	if err != nil || !found {
		return nil
	}
	return spec
}
