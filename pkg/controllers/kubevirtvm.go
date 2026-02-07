package controller

import (
	"context"
	"strings"

	"fmt"

	"github.com/hicompute/kloudstack/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/klog/v2"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

func (c *Controller) OnAddKubevirtVM(req any, isInInitialList bool) {
	ctx := context.Background()
	obj := req.(*unstructured.Unstructured)

	key := types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	var vm kubevirtv1.VirtualMachine

	if err := c.Client.Get(ctx, key, &vm); err != nil {
		klog.Errorf("%v", err)
		return
	}
	if err := c.createVMPasswordCredetialSecret(&vm); err != nil {
		klog.Errorf("%v", err)
		return
	}
}

func (c *Controller) OnDeleteKubevirtVM(req any) {
	ctx := context.Background()
	obj := req.(*unstructured.Unstructured)
	namespace := obj.GetNamespace()
	name := obj.GetName()
	if err := c.removePasswordCredentials(ctx, obj.GetNamespace(), obj.GetName()); err != nil {
		klog.Errorf("%v", err)
	}
	if err := c.updateClusterIP(ctx, namespace, name); err != nil {
		klog.Errorf("%v", err)
	}
}

func (c *Controller) createVMPasswordCredetialSecret(vm *kubevirtv1.VirtualMachine) error {
	labels := vm.GetLabels()
	osCategory := labels["kloudstack/vm/oscategory"]
	if strings.ToLower(osCategory) == "mikrotik" {
		return nil
	}
	ctx := context.Background()
	vmCredentialsSecret := &corev1.Secret{}
	vmCredentialsSecret.Name = fmt.Sprintf("%s-credentials", vm.GetName())
	vmCredentialsSecret.Namespace = vm.GetNamespace()

	username := "root"

	if strings.ToLower(osCategory) == "windows" {
		username = "administrator"
	}

	password := []byte(c.Faker.Internet().Password())
	vmCredentialsSecret.Data = map[string][]byte{}
	vmCredentialsSecret.Data[fmt.Sprintf("%s", username)] = password
	err := c.Client.Create(ctx, vmCredentialsSecret)
	if (err != nil) && (!errors.IsAlreadyExists(err)) {
		return err
	}
	patch := client.MergeFrom(vm.DeepCopy())
	vm.Spec.Template.Spec.AccessCredentials = []kubevirtv1.AccessCredential{
		{
			UserPassword: &kubevirtv1.UserPasswordAccessCredential{
				Source: kubevirtv1.UserPasswordAccessCredentialSource{
					Secret: &kubevirtv1.AccessCredentialSecretSource{
						SecretName: vmCredentialsSecret.Name,
					},
				},
				PropagationMethod: kubevirtv1.UserPasswordAccessCredentialPropagationMethod{
					QemuGuestAgent: &kubevirtv1.QemuGuestAgentUserPasswordAccessCredentialPropagation{},
				},
			},
		},
	}
	if err := c.Client.Patch(ctx, vm, patch); err != nil {
		return err
	}
	return nil
}

func (c *Controller) removePasswordCredentials(ctx context.Context, namespace, name string) error {
	vmCredentials := corev1.Secret{}
	vmCredentials.Name = fmt.Sprintf("%s-credentials", name)
	vmCredentials.Namespace = namespace

	return c.Client.Delete(ctx, &vmCredentials)
}

func (c *Controller) updateClusterIP(ctx context.Context, namespace, name string) error {
	clusterIPList := v1alpha1.ClusterIPList{}
	if err := c.Client.List(ctx, &clusterIPList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.resource", namespace+"/"+name),
		Limit:         1000,
	}); err != nil {
		return err
	}

	for _, clusterIP := range clusterIPList.Items {
		updatedClusterIP := clusterIP.DeepCopy()
		updatedClusterIP.Spec.Mac = ""
		updatedClusterIP.Spec.Interface = ""
		updatedClusterIP.Spec.Resource = ""
		updatedClusterIP.Status.History = append(updatedClusterIP.Status.History,
			v1alpha1.ClusterIPHistory{
				Mac:         clusterIP.Spec.Mac,
				Resource:    namespace + "/" + name,
				AllocatedAt: *clusterIP.Status.History[len(clusterIP.Status.History)-1].AllocatedAt.DeepCopy(),
				ReleasedAt:  metav1.Now(),
				Interface:   clusterIP.Spec.Interface,
			},
		)
		if err := c.Client.Status().Update(ctx, updatedClusterIP); err != nil {
			return err
		}
	}

	return nil
}
