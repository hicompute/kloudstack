package controller

import (
	"github.com/hicompute/kloudstack/pkg/k8s"
	"github.com/jaswdr/faker/v2"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	Client client.Client
	Faker  faker.Faker
}

func New() *Controller {
	client, err := k8s.NewClient()
	if err != nil {
		klog.Fatalf("Error on creating client : %v", err)
	}
	fake := faker.New()
	return &Controller{
		Client: client,
		Faker:  fake,
	}
}
