package scheduler

import (
	"errors"
	"sync"

	mgr "github.com/t-bfame/diago/internal/manager"
	"github.com/t-bfame/diago/pkg/utils"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const hashSize = 6

// PodGroup indicates kind of pod
type PodGroup struct {
	group         string
	clientset     *kubernetes.Clientset
	instanceCount int

	scheduledPods     map[InstanceID]chan Outgoing
	currentCapacities map[InstanceID]uint64
	podmux            sync.Mutex

	outputChannels map[mgr.JobID]chan Event
	workloadCount  map[mgr.JobID]uint32

	qmux     sync.Mutex
	jobQueue *[]mgr.Job
}

// Calculate the number of instances that should be spun up
// TODO: Maxes out at a certain number
func (pg *PodGroup) calculateInstanceCount(frequency uint64, capacity uint64) (count int) {

	// All pods running right now can satisfy capacity
	if frequency <= pg.currentCapacity() {
		return 0
	}

	remaining := frequency - pg.currentCapacity()
	rdr := remaining % capacity

	if rdr == 0 {
		count = int(remaining / capacity)
	} else {
		count = int(remaining/capacity) + 1
	}

	return count
}

// Instances are 0 indexed
func (pg *PodGroup) addInstances(frequency uint64) (err error) {
	pg.podmux.Lock()
	defer pg.podmux.Unlock()

	capacity, err := getCapacity(pg.group)
	if err != nil {
		log.WithField("group", pg.group).WithError(err).Error("Unable to add instances for pod group")
		return err
	}

	count := pg.calculateInstanceCount(frequency, capacity)

	if count == 0 {
		return nil
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
			pod, err := createPodConfig(pg.group, id)
			if err != nil {
				log.WithField("group", pg.group).WithError(err).Error("Unable to add instances for pod group")
				wgErr <- err
				return
			}

			// TODO: Use namespace from env variables
			result, err := pg.clientset.CoreV1().Pods("default").Create(pod)
			if err != nil {
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
		pg.instanceCount += count
	}

	return nil
}

func (pg *PodGroup) removeInstance(instance InstanceID) (err error) {
	pg.podmux.Lock()
	defer pg.podmux.Unlock()

	name := pg.group + "-" + string(instance)

	deletePolicy := metav1.DeletePropagationForeground

	// TODO: Use namespace from env variable
	// Add listeners for detecting deletion
	if err := pg.clientset.CoreV1().Pods("default").Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		return err
	}

	log.WithField("podName", name).WithField("podGroup", pg.group).Info("Removed pod")
	delete(pg.scheduledPods, instance)
	delete(pg.currentCapacities, instance)

	pg.instanceCount--

	return nil
}

func (pg *PodGroup) addJob(j mgr.Job, events chan Event) (err error) {
	// Distribute defer should be stacked on top of unlock
	defer pg.distribute()

	pg.qmux.Lock()
	defer pg.qmux.Unlock()

	pg.outputChannels[j.ID] = events

	// Queue job
	*pg.jobQueue = append(*pg.jobQueue, j)

	return nil
}

func (pg *PodGroup) removeChannel(id mgr.JobID) (err error) {
	events, ok := pg.outputChannels[id]

	if !ok {
		return errors.New("PodGroup does not contain specified JobID")
	}

	delete(pg.outputChannels, id)
	close(events)

	return nil
}

func (pg *PodGroup) registerPod(group string, instance InstanceID) (leader chan Incoming, worker chan Outgoing, err error) {
	defer pg.distribute()
	pg.podmux.Lock()
	defer pg.podmux.Unlock()

	leader = make(chan Incoming) // messages for leader
	worker = make(chan Outgoing) // messages for worker

	capacity, _ := getCapacity(group)

	pg.scheduledPods[instance] = worker
	pg.currentCapacities[instance] = capacity

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

				// Since no more remaining workloads, output channel can be closed
				if pg.workloadCount[jobID] == 0 {
					go pg.distribute()
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

	return leader, worker, nil
}

func (pg *PodGroup) currentCapacity() uint64 {
	var sum uint64 = 0
	for _, freq := range pg.currentCapacities {
		sum += freq
	}

	return sum
}

// TODO: for every jobID, keep track of which pod has been assigned what frequency
// and cleanup this information when finish event from that pod is received
func (pg *PodGroup) distribute() {
	pg.qmux.Lock()
	defer pg.qmux.Unlock()

	j := (*pg.jobQueue)[0]
	frequency := j.Frequency

	// Cannot start since there is not enough capacity
	if frequency > pg.currentCapacity() {
		return
	}

	// Remove next job from queue
	(*pg.jobQueue) = (*pg.jobQueue)[1:]

	for instance, out := range pg.scheduledPods {

		if pg.currentCapacities[instance] == 0 {
			continue
		}

		var workload uint64

		if frequency < pg.currentCapacities[instance] {
			workload = frequency
			pg.currentCapacities[instance] -= frequency
			frequency = 0
		} else {
			workload = pg.currentCapacities[instance]
			frequency -= pg.currentCapacities[instance]
			pg.currentCapacities[instance] = 0
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

	// Send start event on events channel
	output, ok := pg.outputChannels[j.ID]
	if !ok {
		log.WithField("jobID", j.ID).Error("Could not find registered channel for job, discarding start event")
		return
	}

	output <- Start{
		ID:         j.ID,
		Frequency:  j.Frequency,
		Duration:   j.Duration,
		HTTPMethod: j.HTTPMethod,
		HTTPUrl:    j.HTTPUrl,
	}
}

// NewPodGroup Allocates a new podGroup
func NewPodGroup(group string, clientset *kubernetes.Clientset) (pg *PodGroup) {
	pg = new(PodGroup)

	pg.clientset = clientset
	pg.group = group
	pg.instanceCount = 0

	pg.scheduledPods = make(map[InstanceID]chan Outgoing)
	pg.currentCapacities = make(map[InstanceID]uint64)

	pg.outputChannels = make(map[mgr.JobID]chan Event)
	pg.workloadCount = make(map[mgr.JobID]uint32)

	pg.jobQueue = new([]mgr.Job)

	return pg
}
