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

func (pm PodManager) schedule(ti mgr.TestInstance) (id int, err error) {
	groupName := ti.Name + "-" + ti.Id
	var pg PodGroup

	if pg, ok := pm.podGroups[groupName]; !ok {
		pg = NewPodGroup(ti)
		pm.podGroups[groupName] = pg

	}

	return pg.addInstance(pm.clientset)
}

func (pm PodManager) unschedule(ti mgr.TestInstance, instance int) (err error) {
	groupName := ti.Name + "-" + ti.Id
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
