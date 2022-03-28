package scheduler

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	m "github.com/t-bfame/diago/pkg/model"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Scheduler stores data belonging to a scheduler.
type Scheduler struct {
	clientset *kubernetes.Clientset
	model     *SchedulerModel
	podGroups map[string]*PodGroup

	pgmux sync.Mutex
}

// Internal function used to create a pod group in a Scheduler for a group
func (s *Scheduler) createPodGroup(groupName string, failNonExistentGroup bool) (pg *PodGroup, err error) {
	pg, ok := s.podGroups[groupName]

	if ok {
		return pg, nil
	}

	// Pod group creation has to be an atomic operation
	s.pgmux.Lock()
	defer s.pgmux.Unlock()

	cleanupChannel := make(chan struct{})
	pgroup, err := NewPodGroup(groupName, s.clientset, s.model, cleanupChannel, failNonExistentGroup)

	if failNonExistentGroup && err != nil {
		return nil, err
	}

	log.WithField("group", groupName).Debug("Group doesnt exist, creating a new group")
	s.podGroups[groupName] = pgroup

	go func() {
		<-cleanupChannel
		s.pgmux.Lock()
		defer s.pgmux.Unlock()

		log.WithField("group", groupName).Debug("Group does not have remaining workers, removing group instance")

		// Pod group deletion has to be an atomic operation
		delete(s.podGroups, groupName)
	}()

	return s.podGroups[groupName], nil
}

// Submit submits a job in a Scheduler
func (s *Scheduler) Submit(j m.Job) (events chan Event, testId string, instanceId string, err error) {
	events = make(chan Event, 2)

	// If WorkerGroup does not exist while submitting a Job
	// then Job is orphaned and cannot make progress therefore
	// call must fail
	pg, err := s.createPodGroup(j.Group, true)

	pg.testId = testId
	pg.instanceId = instanceId

	if err != nil {
		return nil, "", "", err
	}

	// Add channel for receiving events
	pg.addJob(j, m.TestInstanceID(pg.instanceId), m.TestID(pg.testId), events)

	return events, pg.testId, pg.instanceId, err
}

// Stop stops a job in a Scheduler
func (s *Scheduler) Stop(j m.Job) (err error) {
	groupName := j.Group
	pg, ok := s.podGroups[groupName]

	if !ok {
		return errors.New("Could not find specified groupName")
	}

	pg.removeJob(j.ID)

	return nil
}

// Register registers a WorkerGroup as a PopGroup to the Scheduler.
func (s *Scheduler) Register(group string, instance InstanceID, frequency uint64) (chan Incoming, chan Outgoing, error) {
	// If WorkerGroup does not exist while registration
	// the worker must have been created dynamically
	// and may not have a WorkerGroup in K8s
	pg, _ := s.createPodGroup(group, false)

	// Add test channel for multiplexing
	return pg.registerPod(group, instance, frequency)
}

func (s *Scheduler) GetPgChan(group string, instance InstanceID, frequency uint64) (chan Incoming, chan Outgoing, error) {
	pg, ok := s.podGroups[group]

	if !ok {
		return errors.New("Could not find specified group")
	}

	// Add test channel for multiplexing
	return pg.getInputChan(group, instance, frequency)
}

// NewScheduler creates a new scheduler using in cluster config.
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
