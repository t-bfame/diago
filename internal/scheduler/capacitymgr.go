package scheduler

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
)

// CapacityManager Data structure for keeping track of worker capacities
type CapacityManager struct {
	instanceCount        uint64
	capmux               sync.Mutex
	currentCapacities    map[InstanceID]uint64
	workloadDistribution map[InstanceID]*map[mgr.JobID]uint64
}

// Calculate the number of instances that should be spun up
// TODO: Maxes out at a certain number
func (cm *CapacityManager) calculateInstanceCount(group string, frequency uint64) (int, error) {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	capacity, err := getCapacity(group)
	currentCapacity := capacity * cm.instanceCount

	if err != nil {
		return 0, err
	}

	// All pods running right now can satisfy capacity
	if frequency <= currentCapacity {
		return 0, errors.New("No new pods are required, capacity can be fulfilled")
	}

	remaining := frequency - currentCapacity
	rdr := remaining % capacity
	var count uint64 = 0

	if rdr == 0 {
		count = (remaining / capacity)
	} else {
		count = (remaining / capacity) + 1
	}

	cm.instanceCount += count
	return int(count), nil
}

func (cm *CapacityManager) assignCapacity(instance InstanceID, jobID mgr.JobID, required uint64) (uint64, uint64, error) {
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
	return workload, required, nil
}

func (cm *CapacityManager) reclaimCapacity(instance InstanceID, jobID mgr.JobID) error {
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
	return nil
}

func (cm *CapacityManager) currentCapacity() uint64 {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	var sum uint64 = 0
	for _, freq := range cm.currentCapacities {
		sum += freq
	}

	return sum
}

func (cm *CapacityManager) removeInstance(instance InstanceID) {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	delete(cm.currentCapacities, instance)
	delete(cm.workloadDistribution, instance)
	cm.instanceCount--
}

func (cm *CapacityManager) addInstance(group string, instance InstanceID) error {
	cm.capmux.Lock()
	defer cm.capmux.Unlock()

	capacity, err := getCapacity(group)
	if err != nil {
		return err
	}

	workloadDistribution := make(map[mgr.JobID]uint64)

	cm.currentCapacities[instance] = capacity
	cm.workloadDistribution[instance] = &workloadDistribution

	return nil
}

func (cm *CapacityManager) getPodAssignment(jobID mgr.JobID) *[]InstanceID {
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

// NewCapacityManager returns a new capacity manager
func NewCapacityManager() *CapacityManager {
	var capmgr CapacityManager

	capmgr.currentCapacities = make(map[InstanceID]uint64)
	capmgr.workloadDistribution = make(map[InstanceID]*map[mgr.JobID]uint64)

	return &capmgr
}
