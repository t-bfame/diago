package scheduler

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	m "github.com/t-bfame/diago/pkg/model"
)

// CapacityManager Data structure for keeping track of worker capacities
type CapacityManager struct {
	instanceCount        uint64
	capmux               sync.Mutex
	currentCapacities    map[InstanceID]uint64
	maxCapacities        map[InstanceID]uint64
	cumulativeMaxCap     uint64
	workloadDistribution map[InstanceID]*map[m.JobID]uint64
	podMetrics           map[InstanceID]*PodMetrics
	capacity             uint64

	group string
	model *SchedulerModel
}

/**
* Calculate the number of instances that should be spun up
* TODO: Maxes out at a certain number
*
* @param  frequency  the new frequency that needs to be satisfied
* @return  number of new instances needed
 */
func (cm *CapacityManager) calculateInstanceCount(frequency uint64) (int, error) {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	capacity := cm.capacity
	maxCapacity := cm.cumulativeMaxCap

	// All pods running right now can satisfy capacity
	if frequency <= maxCapacity {
		return 0, errors.New("No new pods are required, capacity can be fulfilled")
	}

	// Calculate number of new instances that needs to be spun up to satisfy new frequency
	remaining := frequency - maxCapacity
	rdr := remaining % capacity
	var count uint64 = 0

	if rdr == 0 {
		count = (remaining / capacity)
	} else {
		count = (remaining / capacity) + 1
	}

	return int(count), nil
}

/**
* Assign a new job to a specific instance in the manager. Try to assign to the instance as much job as possible without
* exceeding its max capacity.
*
* @return  amount of job that was successfully assigned
* @return  amount of job that was left unassigned
 */
func (cm *CapacityManager) assignCapacity(instance InstanceID, jobID m.JobID, required uint64) (uint64, uint64, error) {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	workDis := *cm.workloadDistribution[instance]

	// Completely utilized, move on
	if cm.currentCapacities[instance] == 0 {
		return 0, required, nil
	}

	var workload uint64

	if required <= cm.currentCapacities[instance] {
		workload = required
		cm.currentCapacities[instance] -= workload
		required = 0
	} else {
		workload = cm.currentCapacities[instance]
		required -= cm.currentCapacities[instance]
		cm.currentCapacities[instance] = 0
	}

	workDis[jobID] = workload
	cm.podMetrics[instance].updateCurrentCapacity(cm.currentCapacities[instance])
	return workload, required, nil
}

/**
* Remove a job from an instance in the manager and free the corresponding capacity.
 */
func (cm *CapacityManager) reclaimCapacity(instance InstanceID, jobID m.JobID) error {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	workDis := *cm.workloadDistribution[instance]
	capacity, ok := workDis[jobID]

	if !ok {
		log.WithField("jodId", jobID).Error("Unassigned capacity to worker, invalid state")
		return errors.New("Unassigned capacity to worker, invalid state")
	}

	delete(workDis, jobID)
	cm.currentCapacities[instance] += capacity
	cm.podMetrics[instance].updateCurrentCapacity(cm.currentCapacities[instance])
	return nil
}

/**
* Non-blocking version
* @return  the total available capacity of all instances in the manager
 */
func (cm *CapacityManager) nonBlockingCurrentCapacity() uint64 {
	var sum uint64 = 0
	for _, freq := range cm.currentCapacities {
		sum += freq
	}

	return sum
}

/**
* Blocking version
* @return  the total available capacity of all instances in the manager
 */
func (cm *CapacityManager) currentCapacity() uint64 {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	return cm.nonBlockingCurrentCapacity()
}

/**
* Remove an instance from the manager and adjust the capacity counters
 */
func (cm *CapacityManager) removeInstance(instance InstanceID) {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	cm.cumulativeMaxCap -= cm.maxCapacities[instance]
	cm.instanceCount--

	cm.podMetrics[instance].cleanup()

	delete(cm.maxCapacities, instance)
	delete(cm.podMetrics, instance)
	delete(cm.currentCapacities, instance)
	delete(cm.workloadDistribution, instance)
}

/**
* Add a new instance to the manager. Increment the capacity counters by the given capacity.
 */
func (cm *CapacityManager) addInstance(instance InstanceID, capacity uint64) error {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	workloadDistribution := make(map[m.JobID]uint64)

	cm.instanceCount++
	cm.currentCapacities[instance] = capacity
	cm.maxCapacities[instance] = capacity
	cm.cumulativeMaxCap += capacity
	cm.workloadDistribution[instance] = &workloadDistribution

	cm.podMetrics[instance] = NewPodMetrics(cm.group, instance, capacity)

	return nil
}

/**
* Locate the instances that the given job was assigned to.
*
* @return  list of instance ids that the job was assigned to
 */
func (cm *CapacityManager) getPodAssignment(jobID m.JobID) *[]InstanceID {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	var arr []InstanceID

	for ins, dis := range cm.workloadDistribution {
		if _, ok := (*dis)[jobID]; ok {
			arr = append(arr, ins)
		}
	}

	return &arr
}

/**
* NewCapacityManager returns a new capacity manager
 */
func NewCapacityManager(group string, model *SchedulerModel) *CapacityManager {
	var capmgr CapacityManager

	capmgr.model = model
	capmgr.group = group

	capmgr.maxCapacities = make(map[InstanceID]uint64)
	capmgr.currentCapacities = make(map[InstanceID]uint64)
	capmgr.workloadDistribution = make(map[InstanceID]*map[m.JobID]uint64)
	capmgr.podMetrics = make(map[InstanceID]*PodMetrics)

	// Assign capacity based on the given scheduler model and group
	capmgr.capacity, _ = model.getCapacity(group)

	return &capmgr
}
