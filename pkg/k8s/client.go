package k8s

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewClient() (client.Client, error) {
	cfg, err := createConfig()
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: Scheme,
	})
}

func CreateDynamicClient() *dynamic.DynamicClient {

	config, err := createConfig()
	if err != nil {
		klog.Fatalf("Error creating config: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating dynamicClient: %v", err)
	}
	return dynamicClient
}

func createConfig() (*rest.Config, error) {
	configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	_, err := os.Stat(configFile)
	if err != nil {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", configFile)
}
