package controller

import (
	"github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/hicompute/kloudstack/pkg/ovn"
	"github.com/hicompute/kloudstack/pkg/ovs"
	"github.com/jaswdr/faker/v2"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	Client   client.Client
	ovsAgent ovs.OvsAgent
	ovnAgent ovn.OVNagent
	Faker    faker.Faker
}

func New() *Controller {
	client, err := k8s.NewClient()
	if err != nil {
		klog.Fatalf("Error on creating client : %v", err)
	}
	fake := faker.New()

	ovsAgent, err := ovs.CreateOVSagent()
	if err != nil {
		klog.Fatalf("failed to create ovs agent: %v", err)
	}

	ovnAgent, err := ovn.CreateOVNagent("tcp:192.168.12.177:6641")
	if err != nil {
		klog.Fatalf("failed to create ovs agent: %v", err)
	}

	return &Controller{
		Client:   client,
		ovsAgent: *ovsAgent,
		ovnAgent: *ovnAgent,
		Faker:    fake,
	}
}
