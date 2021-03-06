package client

import (
	"github.com/boz/kubetop/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	Clientset() kubernetes.Interface
}

type client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config

	env util.Env
}

func (c *client) Clientset() kubernetes.Interface {
	return c.clientset
}

func NewClient(env util.Env) (Client, error) {
	env = env.ForComponent("backend/client")

	clientset, config, err := getKubeClientset()
	if err != nil {
		return nil, err
	}
	return &client{clientset, config, env}, nil
}

func getKubeClientset() (*kubernetes.Clientset, *rest.Config, error) {
	config, err := getKubeRestConfig()
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return clientset, config, nil
}

func getKubeRestConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, err
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}
