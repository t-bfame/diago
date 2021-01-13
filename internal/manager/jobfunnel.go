package manager

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/t-bfame/diago/api/v1alpha1"
	c "github.com/t-bfame/diago/config"
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
	testType string,
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

// PrepareScheduledTests sets up crons using test schedules
func (jf *JobFunnel) PrepareScheduledTests(
	client *v1alpha1.DiagoV1Alpha1Client,
	tests *map[string]m.Test, // TODO(frank): rm once we have storage
	instances *map[string][]*m.TestInstance,
) {
	// run all existing
	testSchedules, err := client.TestSchedules(c.Diago.DefaultNamespace).GetAll()
	if err != nil {
		log.WithError(err).Error("Failed to prepare scheduled tests")
	}

	cronRunner := cron.New()
	for _, ts := range testSchedules.Items {
		_, exists := (*tests)[ts.Spec.TestID]
		if !exists {
			log.Error(fmt.Sprintf("Failed to scheduled test: Test with ID %s does not exist", ts.Spec.TestID))
			continue
		}

		_, err = cronRunner.AddFunc(
			ts.Spec.CronSpec,
			func() {
				_, err := jf.BeginTest(m.TestID(ts.Spec.TestID), "scheduled", tests, instances)
				if err != nil {
					log.WithError(err).Error("Failed to begin scheduled test")
				}
			},
		)
		if err != nil {
			log.WithError(err).Error("Failed to scheduled test")
		}
	}

	// watch for changes to test schedules

	// lw := cache.NewListWatchFromClient(
	// 	client.RESTClient, "testschedules", c.Diago.DefaultNamespace, fields.Everything(),
	// )
	//
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			// options.FieldSelector = fields.Everything().String()
			// tsl := v1alpha1.TestScheduleList{}
			// err := client.RESTClient.Get().
			// 	Namespace(c.Diago.DefaultNamespace).
			// 	Resource("testschedules").
			// 	VersionedParams(&options, metav1.ParameterCodec).Do().Into(&tsl)
			// return &tsl, err
			return client.TestSchedules(c.Diago.DefaultNamespace).GetAll()
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.Watch = true
			options.FieldSelector = fields.Everything().String()
			// return client.TestSchedules(c.Diago.DefaultNamespace).Watch()
			return client.RESTClient.Get().
				Namespace(c.Diago.DefaultNamespace).
				Resource("testschedules").
				VersionedParams(&options, metav1.ParameterCodec).
				Watch()
		},
	}
	// wgl := v1alpha1.TestScheduleList{}
	// client.RESTClient.Get().Namespace(c.Diago.DefaultNamespace).Resource("testschedules").Do().Into(&wgl)
	// fmt.Printf("%+v\n", wgl)
	// fmt.Printf("%+v\n", err)
	// obj, err := lw.ListFunc(metav1.ListOptions{})
	// fmt.Printf("%+v\n", obj)
	// fmt.Printf("%+v\n", err)
	// o, err := lw.List(metav1.ListOptions{})
	// fmt.Printf("%+v\n", err)
	// var list *metainternalversion.List
	// list = &metainternalversion.List{Items: make([]runtime.Object, 0, 10)}
	// m, err := meta.ListAccessor(o)

	// fmt.Printf("%+v\n", m)
	// fmt.Printf("%+v\n", list)
	// fmt.Println("asdf")

	_, controller := cache.NewInformer(
		lw, &v1alpha1.TestSchedule{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				spec := obj.(*v1alpha1.TestSchedule).Spec
				log.Info(fmt.Sprintf("TS added. spec: %+v \n", spec))
			},
			DeleteFunc: func(obj interface{}) {
				log.Info(fmt.Sprintf("TS deleted. value: %+v \n", obj))
			},
			UpdateFunc: func(oldObj, obj interface{}) {
				spec := obj.(*v1alpha1.TestSchedule).Spec
				log.Info(fmt.Sprintf("TS updated. spec: %+v \n", spec))
			},
		},
	)
	stop := make(chan struct{})
	controller.Run(stop)

	// store := client.TestSchedules(c.Diago.DefaultNamespace).ListWatch()
	// log.Info(fmt.Sprintf("%+v", store.ListKeys()))

	// watcher, err := client.TestSchedules(c.Diago.DefaultNamespace).Watch()
	// if err != nil {
	// 	log.WithError(err).Error("Failed to watch testschedules")
	// }

	// for e := range watcher.ResultChan() {
	// 	switch e.Type {
	// 	case watch.Added:
	// 		log.Info(fmt.Sprintf("Added %v", e.Object))
	// 	case watch.Modified:
	// 		break
	// 	case watch.Deleted:
	// 		break
	// 	default:
	// 		log.Info(fmt.Sprintf("Added %v", e.Object))
	// 	}
	// }
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
