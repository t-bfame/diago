package v1alpha1

import (
	"k8s.io/client-go/rest"
)

func (c *DiagoV1Alpha1Client) Workers(namespace string) WorkerInterface {
	return &workerClient{
		client: c.restClient,
		ns:     namespace,
	}
}

type DiagoV1Alpha1Client struct {
	restClient rest.Interface
}
