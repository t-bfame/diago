package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type WorkerSpec struct {
	Image                   string `json:"image"`
	Capacity                int    `json:"capacity"`
	AllowedInactivityPeriod int    `json:"allowedInactivityPeriod"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Worker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WorkerSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Worker `json:"items"`
}

type WorkerInterface interface {
	Create(obj *Worker) (*Worker, error)
	Update(obj *Worker) (*Worker, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string) (*Worker, error)
}

type workerClient struct {
	client rest.Interface
	ns     string
}

func (c *workerClient) Create(obj *Worker) (*Worker, error) {
	result := &Worker{}
	err := c.client.Post().
		Namespace(c.ns).Resource("workers").
		Body(obj).Do().Into(result)
	return result, err
}

func (c *workerClient) Update(obj *Worker) (*Worker, error) {
	result := &Worker{}
	err := c.client.Put().
		Namespace(c.ns).Resource("workers").
		Body(obj).Do().Into(result)
	return result, err
}

func (c *workerClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).Resource("workers").
		Name(name).Body(options).Do().
		Error()
}

func (c *workerClient) Get(name string) (*Worker, error) {
	result := &Worker{}
	err := c.client.Get().
		Namespace(c.ns).Resource("workers").
		Name(name).Do().Into(result)
	return result, err
}
