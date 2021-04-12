package manager

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	cm "github.com/t-bfame/diago/pkg/chaosmgr"
	"github.com/t-bfame/diago/pkg/metrics"
	m "github.com/t-bfame/diago/pkg/model"
	s "github.com/t-bfame/diago/pkg/scheduler"
	sto "github.com/t-bfame/diago/pkg/storage"
	"github.com/t-bfame/diago/pkg/tools"
)

// JobFunnel is used to interface with the Scheduler while
// keeping track of ongoing Tests
type JobFunnel interface {
	startOp(key string)
	endOp(key string)
	BeginTest(testID m.TestID, testType string) error
	StopTest(testID m.TestID) error
}

type JobFunnelImpl struct {
	globalLock *sync.Mutex
	testLocks  map[string]*sync.Mutex
	ongoing    map[string]bool
	scheduler  *s.Scheduler
	chaosmgr   *cm.ChaosManager
}

func (jf *JobFunnelImpl) startOp(key string) {
	jf.globalLock.Lock()
	_, exists := jf.testLocks[key]
	if !exists {
		jf.testLocks[key] = &sync.Mutex{}
	}
	jf.globalLock.Unlock()
	jf.testLocks[key].Lock()
}

func (jf *JobFunnelImpl) endOp(key string) {
	jf.testLocks[key].Unlock()
}

func (jf *JobFunnelImpl) RunChaosSimulation(testID m.TestID, chaosInstances []m.ChaosInstance, testDuration uint64) map[m.ChaosID]m.ChaosResult {
	result := make(map[m.ChaosID]m.ChaosResult)
	chaosGroup := sync.WaitGroup{}

	for _, c := range chaosInstances {

		chaosch, deletedPodNames, err := jf.chaosmgr.Simulate(testID, &c, testDuration)

		if err != nil {
			log.WithError(err).WithField("chaosInstance", c.ID).Error("Unable to simulate chaos for instance")
			result[c.ID] = m.ChaosResult{
				Status: m.ChaosFail,
				Error:  err.Error(),
			}
			continue
		}

		chaosGroup.Add(1)

		go func(id m.ChaosID, deletedPodNames []string) {
			defer chaosGroup.Done()
			err := <-chaosch

			if err != nil {
				log.WithError(err).WithField("chaosInstance", id).Error("Chaos simulation failed")
				result[id] = m.ChaosResult{
					Status: m.ChaosFail,
					Error:  err.Error(),
				}
				return
			}

			result[id] = m.ChaosResult{
				Status:      m.ChaosSuccess,
				DeletedPods: deletedPodNames,
				Error:       "",
			}
		}(c.ID, deletedPodNames)
	}

	chaosGroup.Wait()
	return result
}

// BeginTest creates a TestInstance for the Test with the specified TestID
// if another instance of the same Test is not already ongoing
func (jf *JobFunnelImpl) BeginTest(testID m.TestID, testType string) error {
	key := string(testID)
	jf.startOp(key)
	defer jf.endOp(key)

	test, err := sto.GetTestByTestId(m.TestID(key))
	if err != nil || test == nil {
		return fmt.Errorf("Cannot retrieve Test<%s>", key)
	}

	if jf.ongoing[key] {
		return fmt.Errorf("Test<%s> is already ongoing", testID)
	}

	// make instance
	now := time.Now().Unix()
	instanceid := test.Name + "-" + strconv.FormatInt(now, 10)
	instance := &m.TestInstance{
		ID:        m.TestInstanceID(instanceid),
		TestID:    m.TestID(testID),
		Type:      testType,
		Status:    "submitted",
		CreatedAt: now,
	}

	// save instance
	err = sto.AddTestInstance(instance)
	if err != nil {
		return err
	}

	jobGroup := sync.WaitGroup{}
	jobMAggs := map[string]*metrics.Metrics{}
	jobGroupStart := sync.WaitGroup{}
	var testDuration uint64 = 0

	for i, v := range test.Jobs {
		testDuration = tools.Max(testDuration, v.Duration)

		// attempt to submit jobs to scheduler
		ch, err := jf.scheduler.Submit(v)
		if err != nil {
			instance.Status = "failed"
			instance.Error = err.Error()
			// save instance
			sto.AddTestInstance(instance)

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
			return fmt.Errorf("Job<%s> failed to submit: %s", v.ID, err)
		}

		jobGroup.Add(1)
		jobGroupStart.Add(1)

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
					jobGroupStart.Done()
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

	// wait for all jobs to finish or stop
	go func() {
		// Wait for jobs to start
		jobGroupStart.Wait()

		// Complete Chaos simulation with result
		chaosResult := jf.RunChaosSimulation(testID, test.Chaos, testDuration)

		jobGroup.Wait()
		jf.startOp(key)
		defer jf.endOp(key)

		// refresh instance
		instance, err = sto.GetTestInstance(instance.ID)
		if err != nil {
			log.WithField("TestInstanceID", instance.ID).Error(err)
		}

		// If we haven't already stopped this test instance
		if err != nil || !instance.IsTerminal() {
			// save instance
			instance.Status = "done"
			instance.Metrics = jobMAggs
			instance.ChaosResult = chaosResult
			sto.AddTestInstance(instance)
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
	sto.AddTestInstance(instance)

	log.
		WithField("TestID", testID).
		WithField("TestInstanceID", instance.ID).
		Info("Test submitted")
	return nil
}

// StopTest stops the running TestInstance for the Test corresponding
// to the given TestID, if it exists
func (jf *JobFunnelImpl) StopTest(testID m.TestID) error {
	key := string(testID)
	jf.startOp(key)
	defer jf.endOp(key)

	test, err := sto.GetTestByTestId(m.TestID(key))
	if err != nil || test == nil {
		return fmt.Errorf("Cannot retrieve Test<%s>", key)
	}

	if !jf.ongoing[key] {
		return fmt.Errorf("No instance of Test<%s> is currently ongoing", testID)
	}

	for _, c := range test.Chaos {
		jf.chaosmgr.Stop(testID, c.ID)
	}

	for _, v := range test.Jobs {
		err := jf.scheduler.Stop(v)
		if err != nil {
			return fmt.Errorf("Failed to stop Job<%s>", v.ID)
		}
	}

	// we've stopped the test instance
	delete(jf.ongoing, key)

	instances, err := sto.GetTestInstancesByTestID(m.TestID(key))
	if len(instances) == 0 {
		return fmt.Errorf("No instances found for Test<%s>", testID)
	}
	for _, instance := range instances {
		if !instance.IsTerminal() {
			instance.Status = "stopped"
			sto.AddTestInstance(instance)
		}
	}

	log.
		WithField("TestID", testID).
		Info("Test stopped")
	return nil
}

// NewJobFunnel creates a new JobFunnel
func NewJobFunnel(scheduler *s.Scheduler, cm *cm.ChaosManager) JobFunnel {
	jf := &JobFunnelImpl{
		&sync.Mutex{},
		map[string]*sync.Mutex{},
		map[string]bool{},
		scheduler,
		cm,
	}
	return jf
}

type TestingJobFunnel struct {
	Starts []m.TestID
	Stops  []m.TestID
}

func (jf *TestingJobFunnel) startOp(key string) {}
func (jf *TestingJobFunnel) endOp(key string)   {}
func (jf *TestingJobFunnel) BeginTest(
	testID m.TestID,
	testType string,
) error {
	jf.Starts = append(jf.Starts, testID)
	return nil
}
func (jf *TestingJobFunnel) StopTest(
	testID m.TestID,
) error {
	jf.Stops = append(jf.Stops, testID)
	return nil
}
