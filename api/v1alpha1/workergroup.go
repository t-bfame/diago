package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type WorkerGroupSpec struct {
	Image                   string `json:"image"`
	Capacity                int    `json:"capacity"`
	AllowedInactivityPeriod int    `json:"allowedInactivityPeriod"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WorkerGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WorkerGroupSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WorkerGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []WorkerGroup `json:"items"`
}

type WorkerGroupInterface interface {
	Create(obj *WorkerGroup) (*WorkerGroup, error)
	Update(obj *WorkerGroup) (*WorkerGroup, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string) (*WorkerGroup, error)
}

type workerGroupClient struct {
	client rest.Interface
	ns     string
}

func (c *workerGroupClient) Create(obj *WorkerGroup) (*WorkerGroup, error) {
	result := &WorkerGroup{}
	err := c.client.Post().
		Namespace(c.ns).Resource("workergroups").
		Body(obj).Do().Into(result)
	return result, err
}

func (c *workerGroupClient) Update(obj *WorkerGroup) (*WorkerGroup, error) {
	result := &WorkerGroup{}
	err := c.client.Put().
		Namespace(c.ns).Resource("workergroups").
		Body(obj).Do().Into(result)
	return result, err
}

func (c *workerGroupClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).Resource("workergroups").
		Name(name).Body(options).Do().
		Error()
}

func (c *workerGroupClient) Get(name string) (*WorkerGroup, error) {
	result := &WorkerGroup{}
	err := c.client.Get().
		Namespace(c.ns).Resource("workergroups").
		Name(name).Do().Into(result)
	return result, err
}
