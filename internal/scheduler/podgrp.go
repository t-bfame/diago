package scheduler

import (
	"errors"

	mgr "github.com/t-bfame/diago/internal/manager"
	utils "github.com/t-bfame/diago/pkg/utils"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const hashSize = 6

// PodGroup indicates kind of pod
type PodGroup struct {
	group         string
	clientset     *kubernetes.Clientset
	scheduledPods map[InstanceID]chan Outgoing
	instanceCount int

	outputChannels map[mgr.JobID]chan Event
}

// Instances are 0 indexed
func (pg PodGroup) addInstance() (instance InstanceID, err error) {
	id := InstanceID(utils.RandHash(hashSize))

	pod, err := createPodConfig(pg.group, id)

	if err != nil {
		return InstanceID(""), err
	}

	result, err := pg.clientset.CoreV1().Pods("default").Create(pod)
	if err != nil {
		return InstanceID(""), err
	}

	log.WithField("podName", result.GetObjectMeta().GetName()).WithField("podGroup", pg.group).Info("Created pod")
	pg.instanceCount++

	return id, nil
}

func (pg PodGroup) removeInstance(instance InstanceID) (err error) {
	name := pg.group + "-" + string(instance)

	deletePolicy := metav1.DeletePropagationForeground
	if err := pg.clientset.CoreV1().Pods("default").Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return err
	}

	log.WithField("podName", name).WithField("podGroup", pg.group).Info("Removed pod")
	delete(pg.scheduledPods, instance)
	pg.instanceCount--

	return nil
}

func (pg PodGroup) addChannel(id mgr.JobID, events chan Event) (err error) {
	pg.outputChannels[id] = events

	return nil
}

func (pg PodGroup) removeChannel(id mgr.JobID) (err error) {
	events, ok := pg.outputChannels[id]

	if !ok {
		return errors.New("PodGroup does not contain specified JobID")
	}

	delete(pg.outputChannels, id)
	close(events)

	return nil
}

func (pg PodGroup) registerPod(instance InstanceID) (leader chan Incoming, worker chan Outgoing, err error) {
	leader = make(chan Incoming) // messages for leader
	worker = make(chan Outgoing) // messages for worker

	pg.scheduledPods[instance] = worker

	// Mux events from pod to correct job channels
	go func() {
		for msg := range leader {

			jobID := msg.getJobID()

			output, ok := pg.outputChannels[jobID]
			if !ok {
				log.WithField("jobID", jobID).Error("Could not find registered channel for job, discarding event")
				continue
			}

			// Send event to respective output channel
			output <- msg
		}

		// If channel is closed then communication with pod has stopped
		pg.removeInstance(instance)
	}()

	return leader, worker, nil
}

// NewPodGroup Allocates a new podGroup
func NewPodGroup(group string, clientset *kubernetes.Clientset) (pg *PodGroup) {
	pg = new(PodGroup)

	pg.clientset = clientset
	pg.group = group
	pg.scheduledPods = make(map[InstanceID]chan Outgoing)
	pg.instanceCount = 0

	pg.outputChannels = make(map[mgr.JobID]chan Event)

	return pg
}
