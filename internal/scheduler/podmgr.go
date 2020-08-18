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

func (pm PodManager) register(group string, instance InstanceID) (leader chan Incoming, worker chan Outgoing, err error) {
	pg, ok := pm.podGroups[group]

	if !ok {
		return nil, nil, errors.New("Could not find specified group")
	}

	// Add test channel for multiplexing
	return pg.registerPod(group, instance)
}

func (pm PodManager) schedule(j mgr.Job, events chan Event) (err error) {
	groupName := j.Group
	pg, ok := pm.podGroups[groupName]

	if !ok {
		pg = NewPodGroup(groupName, pm.clientset)
		pm.podGroups[groupName] = pg
	}

	// Add channel for receiving events
	pg.addJob(j, events)

	if j.Frequency > pg.currentCapacity() {
		defer pg.addInstances(j.Frequency)
	}

	return nil
}

func (pm PodManager) unschedule(j mgr.Job) (err error) {
	groupName := j.Group
	pg, ok := pm.podGroups[groupName]

	if !ok {
		return errors.New("Could not find specified groupName")
	}

	pg.removeChannel(j.ID)

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

	return pm
}
