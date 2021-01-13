package v1alpha1

import (
	"k8s.io/client-go/rest"
)

func (c *DiagoV1Alpha1Client) WorkerGroups(namespace string) WorkerGroupInterface {
	return &workerGroupClient{
		client: c.RESTClient,
		ns:     namespace,
	}
}

func (c *DiagoV1Alpha1Client) TestSchedules(namespace string) TestScheduleInterface {
	return &testScheduleClient{
		client: c.RESTClient,
		ns:     namespace,
	}
}

type DiagoV1Alpha1Client struct {
	RESTClient *rest.RESTClient
}
