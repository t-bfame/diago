package scheduler

import (
	"errors"

	log "github.com/sirupsen/logrus"
	m "github.com/t-bfame/diago/internal/model"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Scheduler lalalal
type Scheduler struct {
	clientset *kubernetes.Clientset
	model     *SchedulerModel
	podGroups map[string]*PodGroup
}

func (s Scheduler) createPodGroup(groupName string) (pg *PodGroup) {
	s.podGroups[groupName] = NewPodGroup(groupName, s.clientset, s.model)

	return s.podGroups[groupName]
}

// Submit dingdingi
func (s Scheduler) Submit(j m.Job) (events chan Event, err error) {
	events = make(chan Event, 2)

	group := j.Group
	pg, ok := s.podGroups[group]

	if !ok {
		log.WithField("group", group).Debug("Group doesnt exist, creating a new group")
		pg = s.createPodGroup(group)
	}

	// Add channel for receiving events
	pg.addJob(j, events)

	return events, err
}

// Stop dongdongdong
func (s Scheduler) Stop(j m.Job) (err error) {
	groupName := j.Group
	pg, ok := s.podGroups[groupName]

	if !ok {
		return errors.New("Could not find specified groupName")
	}

	pg.removeJob(j.ID)

	return nil
}

// Register something
func (s Scheduler) Register(group string, instance InstanceID, frequency uint64) (chan Incoming, chan Outgoing, error) {
	pg, ok := s.podGroups[group]

	// In this case, leader was not responsible for spinning up the worker process
	// and it will simply create initialize a new group to accomodate discovery
	if !ok {
		log.WithField("group", group).Debug("Group doesnt exist, creating a new group during registration")
		pg = s.createPodGroup(group)
	}

	// Add test channel for multiplexing
	return pg.registerPod(group, instance, frequency)
}

// NewScheduler laalala
func NewScheduler() *Scheduler {
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

	s := new(Scheduler)
	s.clientset = clientset
	s.podGroups = make(map[string]*PodGroup)

	s.model, err = NewSchedulerModel(config)

	if err != nil {
		panic(err.Error())
	}

	return s
}
