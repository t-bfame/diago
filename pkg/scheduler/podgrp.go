package scheduler

import (
	"errors"
	"sync"

	c "github.com/t-bfame/diago/config"
	m "github.com/t-bfame/diago/pkg/model"
	"github.com/t-bfame/diago/pkg/utils"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const hashSize = 6

// PodGroup indicates kind of pod
type PodGroup struct {
	group     string
	clientset *kubernetes.Clientset
	model     *SchedulerModel

	scheduledPods map[InstanceID]chan Outgoing
	podmux        sync.Mutex

	outputChannels map[m.JobID]chan Event
	workloadCount  map[m.JobID]uint32

	qmux     sync.Mutex
	jobQueue *[]m.Job

	capmgr *CapacityManager

	cleanupChannel chan struct{}
}

// Instances are 0 indexed
func (pg *PodGroup) addInstances(frequency uint64) (err error) {
	pg.podmux.Lock()
	defer pg.podmux.Unlock()

	count, err := pg.capmgr.calculateInstanceCount(frequency)

	if err != nil {
		log.WithField("group", pg.group).WithError(err).Error("Unable to add instances for pod group")
		return err
	}

	var wg sync.WaitGroup
	wgErr := make(chan error)
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()

			id := InstanceID(utils.RandHash(hashSize))

			// Assumption: Pod configuration is correct and it will always come up
			// TODO: Add listener that listens whether pod was initialized
			pod, err := pg.model.createPodConfig(pg.group, id)
			if err != nil {
				log.WithField("group", pg.group).WithError(err).Error("Unable to add instances for pod group")
				wgErr <- err
				return
			}

			result, err := pg.clientset.CoreV1().Pods(c.Diago.DefaultNamespace).Create(pod)
			if err != nil {
				log.WithField("group", pg.group).WithError(err).Error("Unable to add instances for pod group")
				wgErr <- err
				return
			}

			log.WithField("pod", result.GetObjectMeta().GetName()).WithField("podGroup", pg.group).Info("Created pod")
		}()
	}

	wg.Wait()

	select {
	case <-wgErr:
		return err
	default:
		return nil
	}
}

func (pg *PodGroup) removeInstance(instance InstanceID) (err error) {
	pg.podmux.Lock()
	defer pg.podmux.Unlock()

	name := pg.group + "-" + string(instance)

	delete(pg.scheduledPods, instance)
	pg.capmgr.removeInstance(instance)

	deletePolicy := metav1.DeletePropagationForeground

	// Since there are no more workers remaining we can
	// cleanup the pg instance from the scheduler
	if len(pg.scheduledPods) == 0 {
		close(pg.cleanupChannel)
	}

	// Add listeners for detecting deletion
	if err := pg.clientset.CoreV1().Pods(c.Diago.DefaultNamespace).Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.WithError(err).WithField("podName", name).WithField("podGroup", pg.group).Error("Encountered error while pod deletion")
		return err
	}

	log.WithField("podName", name).WithField("podGroup", pg.group).Info("Removed pod")

	return nil
}

func (pg *PodGroup) addJob(j m.Job, events chan Event) (err error) {
	pg.qmux.Lock()
	defer pg.qmux.Unlock()

	pg.outputChannels[j.ID] = events

	// Queue job
	*pg.jobQueue = append(*pg.jobQueue, j)
	go pg.addInstances(j.Frequency)

	pg.distribute()
	return nil
}

func (pg *PodGroup) removeJob(id m.JobID) (err error) {

	for _, instance := range *(pg.capmgr.getPodAssignment(id)) {
		pg.scheduledPods[instance] <- Stop{id}
	}

	return nil
}

func (pg *PodGroup) registerPod(group string, instance InstanceID, frequency uint64) (leader chan Incoming, worker chan Outgoing, err error) {
	pg.qmux.Lock()
	defer pg.qmux.Unlock()

	leader = make(chan Incoming)    // messages for leader
	worker = make(chan Outgoing, 2) // messages for worker

	pg.scheduledPods[instance] = worker
	pg.capmgr.addInstance(instance, frequency)

	// Mux events from pod to correct job channels
	go func() {
		for msg := range leader {

			jobID := msg.getJobID()

			output, ok := pg.outputChannels[jobID]
			if !ok {
				log.WithField("jobID", jobID).Error("Could not find registered channel for job, discarding event")
				continue
			}

			switch msg.(type) {
			case Finish:
				pg.workloadCount[jobID]--

				pg.capmgr.reclaimCapacity(instance, jobID)
				pg.distribute()

				// Since no more remaining workloads, output channel can be closed
				if pg.workloadCount[jobID] == 0 {
					delete(pg.workloadCount, jobID)
					delete(pg.outputChannels, jobID)
					close(output)
				}

			default:
				// Send event to respective output channel
				output <- msg
			}
		}

		// If channel is closed then communication with pod has stopped
		pg.removeInstance(instance)
	}()

	pg.distribute()
	return leader, worker, nil
}

// TODO: for every jobID, keep track of which pod has been assigned what frequency
// and cleanup this information when finish event from that pod is received
func (pg *PodGroup) distribute() {
	if len(*pg.jobQueue) == 0 {
		return
	}

	j := (*pg.jobQueue)[0]
	frequency := j.Frequency
	var workload uint64
	var err error

	// Cannot start since there is not enough capacity
	if frequency > pg.capmgr.currentCapacity() {
		return
	}

	// Remove next job from queue
	(*pg.jobQueue) = (*pg.jobQueue)[1:]

	for instance, out := range pg.scheduledPods {

		workload, frequency, err = pg.capmgr.assignCapacity(instance, j.ID, frequency)

		if err != nil {
			continue
		}

		// This instance was not used
		if workload == 0 {
			continue
		}

		// Increment the worload count
		pg.workloadCount[j.ID]++
		out <- Start{
			ID:         j.ID,
			Frequency:  workload,
			Duration:   j.Duration,
			HTTPMethod: j.HTTPMethod,
			HTTPUrl:    j.HTTPUrl,
		}

		if frequency == 0 {
			break
		}
	}

	if frequency > 0 {
		log.WithField("jobID", j.ID).Warning("Assigned partial workload, continuing test")
	}

	// Send start event on events channel
	output, ok := pg.outputChannels[j.ID]
	if !ok {
		log.WithField("jobID", j.ID).Error("Could not find registered channel for job, discarding start event")
		return
	}

	output <- Start{
		ID:         j.ID,
		Frequency:  j.Frequency - frequency,
		Duration:   j.Duration,
		HTTPMethod: j.HTTPMethod,
		HTTPUrl:    j.HTTPUrl,
	}
}

// NewPodGroup Allocates a new podGroup
func NewPodGroup(group string, clientset *kubernetes.Clientset, model *SchedulerModel, cleanup chan struct{}, failNonExistentGroup bool) (pg *PodGroup, err error) {

	// If we want to fail when WorkerGroup doesnt exist in K8s
	if failNonExistentGroup && !model.checkExists(group) {
		return nil, errors.New("WorkerGroup does not exist")
	}

	pg = new(PodGroup)

	pg.clientset = clientset
	pg.model = model
	pg.group = group

	pg.scheduledPods = make(map[InstanceID]chan Outgoing)
	pg.outputChannels = make(map[m.JobID]chan Event)
	pg.workloadCount = make(map[m.JobID]uint32)

	pg.jobQueue = new([]m.Job)
	pg.capmgr = NewCapacityManager(group, model)

	pg.cleanupChannel = cleanup

	return pg, nil
}
