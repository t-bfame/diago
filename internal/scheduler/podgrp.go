package scheduler

import (
	"errors"

	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodGroup indicates kind of pod
type PodGroup struct {
	ti            mgr.TestInstance
	scheduledPods map[string]string
	instanceIndex int
	instanceCount int
}

func (pg PodGroup) getLabels() map[string]string {
	group := pg.ti.Name + "-" + pg.ti.Id

	labels := map[string]string{
		"group":    group,
		"instance": string(pg.instanceCount),
	}

	return labels
}

// Instances are 0 indexed
func (pg PodGroup) addInstance(clientset *kubernetes.Clientset) (instance int, err error) {

	group := pg.ti.Name + "-" + pg.ti.Id
	name := group + "-" + string(pg.instanceIndex)
	count := pg.instanceIndex

	pod, err := createPodConfig(name, pg.ti.Image, pg.ti.Env, pg.getLabels())

	if err != nil {
		return 0, err
	}

	log.WithField("podName", name).WithField("podGroup", group).Info("About to create pod")

	result, err := clientset.CoreV1().Pods("default").Create(pod)
	if err != nil {
		return 0, err
	}

	pg.scheduledPods[name] = "created"
	log.WithField("podName", result.GetObjectMeta().GetName()).WithField("podGroup", group).Info("Created pod")
	pg.instanceCount++
	pg.instanceIndex++

	return count, nil
}

func (pg PodGroup) removeInstance(clientset *kubernetes.Clientset, instance int) (err error) {

	if instance >= pg.instanceCount {
		return errors.New("Cannot remove nonexitent instance")
	}

	group := pg.ti.Name + "-" + pg.ti.Id
	name := group + "-" + string(instance)

	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.CoreV1().Pods("default").Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return err
	}

	log.WithField("podName", name).WithField("podGroup", group).Info("Removed pod")
	delete(pg.scheduledPods, name)
	pg.instanceCount--

	return nil
}

// NewPodGroup Allocates a new podGroup
func NewPodGroup(ti mgr.TestInstance) (podGroup *PodGroup) {
	podGroup = new(PodGroup)

	podGroup.ti = ti
	podGroup.scheduledPods = make(map[string]string)
	podGroup.instanceIndex = 0
	podGroup.instanceCount = 0

	return podGroup
}
