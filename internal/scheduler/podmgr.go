package scheduler

import (
	"errors"

	mgr "github.com/t-bfame/diago/internal/manager"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// PodManager Manages pods created by diago in K8s cluster
type PodManager struct {
	clientset *kubernetes.Clientset
	podGroups map[string]*PodGroup
}

func (pm PodManager) register(group string, instance int) (events <-chan Event, err error) {
	pg, ok := pm.podGroups[group]

	if !ok {
		return nil, errors.New("Could not find specified group")
	}

	// Add test channel for multiplexing
	events, err = pg.registerPod(instance)

	return events, err
}

func (pm PodManager) schedule(j mgr.Job, events chan Event) (err error) {
	groupName := j.Group
	var pg PodGroup

	if pg, ok := pm.podGroups[groupName]; !ok {
		pg = NewPodGroup(groupName)
		pm.podGroups[groupName] = pg
	}

	// Add channel for receiving events
	pg.addChannel(j.ID, events)

	// TODO: Check capacity and add instance
	// defer pg.addInstance(pm.clientset)

	return nil
}

func (pm PodManager) unschedule(j mgr.Job, instance int) (err error) {
	groupName := j.Group
	pg, ok := pm.podGroups[groupName]

	if !ok {
		return errors.New("Could not find specified groupName")
	}

	return pg.removeInstance(pm.clientset, instance)
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

	return pm
}
