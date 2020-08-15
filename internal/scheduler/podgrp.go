package scheduler

import (
	"errors"

	mgr "github.com/t-bfame/diago/internal/manager"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodGroup indicates kind of pod
type PodGroup struct {
	group         string
	scheduledPods map[string]string
	instanceIndex int
	instanceCount int

	outputChannels map[mgr.JobID]chan Event
}

// Instances are 0 indexed
func (pg PodGroup) addInstance(clientset *kubernetes.Clientset) (instance int, err error) {
	name := pg.group + "-" + string(pg.instanceIndex)
	count := pg.instanceIndex

	pod, err := createPodConfig(pg.group, count)

	if err != nil {
		return 0, err
	}

	result, err := clientset.CoreV1().Pods("default").Create(pod)
	if err != nil {
		return 0, err
	}

	pg.scheduledPods[name] = "created"
	log.WithField("podName", result.GetObjectMeta().GetName()).WithField("podGroup", pg.group).Info("Created pod")
	pg.instanceCount++
	pg.instanceIndex++

	return count, nil
}

func (pg PodGroup) removeInstance(clientset *kubernetes.Clientset, instance int) (err error) {

	if instance >= pg.instanceCount {
		return errors.New("Cannot remove nonexitent instance")
	}

	name := pg.group + "-" + string(instance)

	deletePolicy := metav1.DeletePropagationForeground
	if err := clientset.CoreV1().Pods("default").Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return err
	}

	log.WithField("podName", name).WithField("podGroup", pg.group).Info("Removed pod")
	delete(pg.scheduledPods, name)
	pg.instanceCount--

	return nil
}

func (pg PodGroup) addChannel(id mgr.JobID, events chan Event) (err error) {
	pg.outputChannels[id] = events

	return nil
}

func (pg PodGroup) registerPod(instance int) (events chan Event, err error) {
	events = make(chan Event)

	// Fan-out events from pod to correct job channel
	go func() {
		for msg := range events {

			jobID := msg.getJobID()

			output, ok := pg.outputChannels[jobID]
			if !ok {
				log.WithField("jobID", jobID).Error("Could not find registered channel for job, discarding event")
				continue
			}

			// Send event to respective output channel
			output <- msg
		}
	}()

	return events, nil
}

// NewPodGroup Allocates a new podGroup
func NewPodGroup(group string) (pg *PodGroup) {
	pg = new(PodGroup)

	pg.group = group
	pg.scheduledPods = make(map[string]string)
	pg.instanceIndex = 0
	pg.instanceCount = 0

	pg.outputChannels = make(map[mgr.JobID]chan Event)

	return pg
}
