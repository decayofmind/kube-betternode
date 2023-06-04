package k8s

import (
	"errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient() (*kubernetes.Clientset, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, errors.New("Couldn't create Kubernetes default config: " + err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.New("Couldn't create Kubernetes client: " + err.Error())
	}

	return client, nil
}
