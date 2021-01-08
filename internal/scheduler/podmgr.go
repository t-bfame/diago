package scheduler

import (
	"errors"

	m "github.com/t-bfame/diago/internal/model"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	log "github.com/sirupsen/logrus"
)

// PodManager Manages pods created by diago in K8s cluster
type PodManager struct {
	clientset *kubernetes.Clientset
	model     *SchedulerModel
	podGroups map[string]*PodGroup
}

func (pm PodManager) createPodGroup(groupName string) (pg *PodGroup) {
	pm.podGroups[groupName] = NewPodGroup(groupName, pm.clientset, pm.model)

	return pm.podGroups[groupName]
}

func (pm PodManager) register(group string, instance InstanceID, frequency uint64) (leader chan Incoming, worker chan Outgoing, err error) {
	pg, ok := pm.podGroups[group]

	// In this case, leader was not responsible for spinning up the worker process
	// and it will simply create initialize a new group to accomodate discovery
	if !ok {
		log.WithField("group", group).Debug("Group doesnt exist, creating a new group during registration")
		pg = pm.createPodGroup(group)
	}

	// Add test channel for multiplexing
	return pg.registerPod(group, instance, frequency)
}

func (pm PodManager) schedule(j m.Job, events chan Event) (err error) {
	group := j.Group
	pg, ok := pm.podGroups[group]

	if !ok {
		log.WithField("group", group).Debug("Group doesnt exist, creating a new group")
		pg = pm.createPodGroup(group)
	}

	// Add channel for receiving events
	pg.addJob(j, events)

	return nil
}

func (pm PodManager) unschedule(j m.Job) (err error) {
	groupName := j.Group
	pg, ok := pm.podGroups[groupName]

	if !ok {
		return errors.New("Could not find specified groupName")
	}

	pg.removeJob(j.ID)

	return nil
}

// NewPodManager Creates a new PodManager
func NewPodManager() *PodManager {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pm := new(PodManager)
	pm.clientset = clientset
	pm.podGroups = make(map[string]*PodGroup)

	pm.model, err = NewSchedulerModel(config)

	if err != nil {
		panic(err.Error())
	}

	return pm
}
