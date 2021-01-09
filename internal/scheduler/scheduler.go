package scheduler

import (
	"errors"
	"sync"

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

	pgmux sync.Mutex
}

func (s *Scheduler) createPodGroup(groupName string) (pg *PodGroup) {
	pg, ok := s.podGroups[groupName]

	if ok {
		return pg
	}

	// Pod group creation has to be an atomic operation
	s.pgmux.Lock()
	defer s.pgmux.Unlock()

	log.WithField("group", groupName).Info("Group doesnt exist, creating a new group")

	cleanupChannel := make(chan struct{})
	s.podGroups[groupName] = NewPodGroup(groupName, s.clientset, s.model, cleanupChannel)

	go func() {
		<-cleanupChannel
		s.pgmux.Lock()
		defer s.pgmux.Unlock()

		log.WithField("group", groupName).Info("Group does not have remaining workers, removing group instance")

		// Pod group deletion has to be an atomic operation
		delete(s.podGroups, groupName)
	}()

	return s.podGroups[groupName]
}

// Submit dingdingi
func (s *Scheduler) Submit(j m.Job) (events chan Event, err error) {
	events = make(chan Event, 2)

	group := j.Group
	pg := s.createPodGroup(group)

	// Add channel for receiving events
	pg.addJob(j, events)

	return events, err
}

// Stop dongdongdong
func (s *Scheduler) Stop(j m.Job) (err error) {
	groupName := j.Group
	pg, ok := s.podGroups[groupName]

	if !ok {
		return errors.New("Could not find specified groupName")
	}

	pg.removeJob(j.ID)

	return nil
}

// Register something
func (s *Scheduler) Register(group string, instance InstanceID, frequency uint64) (chan Incoming, chan Outgoing, error) {
	pg := s.createPodGroup(group)

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
