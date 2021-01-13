package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type TestScheduleSpec struct {
	TestID   string `json:"testID"`
	CronSpec string `json:"cronSpec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TestSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TestScheduleSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TestScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []TestSchedule `json:"items"`
}

type TestScheduleInterface interface {
	GetAll() (*TestScheduleList, error)
	Watch() (watch.Interface, error)
}

type testScheduleClient struct {
	client *rest.RESTClient
	ns     string
}

func (c *testScheduleClient) GetAll() (*TestScheduleList, error) {
	result := &TestScheduleList{}
	err := c.client.Get().
		Namespace(c.ns).Resource("testschedules").
		Do().Into(result)
	return result, err
}

func (c *testScheduleClient) Watch() (watch.Interface, error) {
	return c.client.Get().Namespace(c.ns).
		Resource("testschedules").
		Watch()
}
