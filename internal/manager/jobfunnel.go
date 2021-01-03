package manager

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/t-bfame/diago/internal/metrics"
	m "github.com/t-bfame/diago/internal/model"
	s "github.com/t-bfame/diago/internal/scheduler"
)

// JobFunnel is used to interface with the Scheduler while
// keeping track of ongoing Tests
type JobFunnel struct {
	globalLock *sync.Mutex
	testLocks  map[string]*sync.Mutex
	ongoing    map[string]bool
	scheduler  *s.Scheduler
}

func (jf *JobFunnel) startOp(key string) {
	jf.globalLock.Lock()
	_, exists := jf.testLocks[key]
	if !exists {
		jf.testLocks[key] = &sync.Mutex{}
	}
	jf.globalLock.Unlock()
	jf.testLocks[key].Lock()
}

func (jf *JobFunnel) endOp(key string) {
	jf.testLocks[key].Unlock()
}

// BeginTest creates a TestInstance for the Test with the specified TestID
// if another instance of the same Test is not already ongoing
func (jf *JobFunnel) BeginTest(
	testID m.TestID,
	tests *map[string]m.Test, // TODO(frank): rm once we have storage
	instances *map[string][]*m.TestInstance,
) (bool, error) {
	key := string(testID)
	jf.startOp(key)
	defer jf.endOp(key)

	test, exists := (*tests)[key]
	if !exists {
		return false, fmt.Errorf("Test<%s> does not exist", testID)
	}

	if jf.ongoing[key] {
		return false, fmt.Errorf("Test<%s> is already ongoing", testID)
	}

	// make instance
	now := time.Now().Unix()
	instanceid := test.Name + "-" + strconv.FormatInt(now, 10)
	instance := &m.TestInstance{
		ID:        m.TestInstanceID(instanceid),
		TestID:    m.TestID(testID),
		Type:      "adhoc",
		Status:    "submitted",
		CreatedAt: now,
	}

	// save instance
	(*instances)[key] = append((*instances)[key], instance)

	jobGroup := sync.WaitGroup{}
	jobMAggs := map[string]*metrics.Metrics{}

	for i, v := range test.Jobs {
		// attempt to submit jobs to scheduler
		ch, err := jf.scheduler.Submit(v)
		if err != nil {
			instance.Status = "failed"

			// save instance
			for i, ti := range (*instances)[key] {
				if ti.ID == instance.ID {
					(*instances)[key][i] = instance
				}
			}

			// cancel previously submitted jobs
			for idx := 0; idx < i; idx++ {
				err := jf.scheduler.Stop(v)
				if err != nil {
					// bummer...
					log.
						WithField("TestID", testID).
						WithField("TestInstanceID", instance.ID).
						WithField("JobID", v.ID).
						Info("Failed to stop job")
				}
			}
			return false, fmt.Errorf("Job<%s> failed to submit: %s", v.ID, err)
		}

		jobGroup.Add(1)
		mAgg := metrics.NewMetricAggregator(
			string(testID),
			string(instance.ID),
			string(v.ID),
		)
		jobMAggs[string(v.ID)] = mAgg

		// listen on each channel for job events
		go func(j m.Job, mAgg *metrics.Metrics) {
			defer jobGroup.Done()
			for msg := range ch {
				switch x := msg.(type) {
				case s.Metrics:
					mAgg.Add(&x)
				case s.Start:
					log.WithField("Start event", msg).Info("Starting job")
				default:
				}
			}
			mAgg.Close()
			log.
				WithField("TestID", testID).
				WithField("TestInstanceID", instance.ID).
				WithField("JobID", j.ID).
				Info("Finished/Stopped Job")
		}(v, mAgg)
	}

	// wait for all jobs to finish
	go func() {
		jobGroup.Wait()
		jf.startOp(key)
		defer jf.endOp(key)

		// instance completed -> save
		instance.Status = "done"
		instance.Metrics = jobMAggs
		for i, ti := range (*instances)[key] {
			if ti.ID == instance.ID {
				(*instances)[key][i] = instance
			}
		}

		delete(jf.ongoing, key)
		log.
			WithField("TestID", testID).
			WithField("TestInstanceID", instance.ID).
			Info("Finished Test")
	}()

	// we've successfully submitted the jobs
	jf.ongoing[key] = true

	instance.Status = "submitted"
	for i, ti := range (*instances)[key] {
		if ti.ID == instance.ID {
			(*instances)[key][i] = instance
		}
	}

	log.
		WithField("TestID", testID).
		WithField("TestInstanceID", instance.ID).
		Info("Test submitted")
	return true, nil
}

// StopTest stops the running TestInstance for the Test corresponding
// to the given TestID, if it exists
func (jf *JobFunnel) StopTest(
	testID m.TestID,
	tests *map[string]m.Test, // TODO(frank): rm once we have storage
	instances *map[string][]*m.TestInstance,
) (bool, error) {
	key := string(testID)
	jf.startOp(key)
	defer jf.endOp(key)

	test, exists := (*tests)[key]
	if !exists {
		return false, fmt.Errorf("Test<%s> does not exist", testID)
	}

	if !jf.ongoing[key] {
		return false, fmt.Errorf("No instance of Test<%s> is currently ongoing", testID)
	}

	for _, v := range test.Jobs {
		err := jf.scheduler.Stop(v)
		if err != nil {
			return false, fmt.Errorf("Failed to stop Job<%s>", v.ID)
		}
	}

	// we've stopped the test instance
	delete(jf.ongoing, key)

	_, exists = (*instances)[key]
	if !exists {
		return false, fmt.Errorf("No instances found for Test<%s>", testID)
	}
	for _, instance := range (*instances)[key] {
		if !instance.IsTerminal() {
			instance.Status = "stopped"
			for i, ti := range (*instances)[key] {
				if ti.ID == instance.ID {
					(*instances)[key][i] = instance
				}
			}
		}
	}

	log.
		WithField("TestID", testID).
		Info("Test stopped")
	return true, nil
}

// NewJobFunnel creates a new JobFunnel
func NewJobFunnel(scheduler *s.Scheduler) *JobFunnel {
	jf := JobFunnel{
		&sync.Mutex{},
		map[string]*sync.Mutex{},
		map[string]bool{},
		scheduler,
	}
	return &jf
}
