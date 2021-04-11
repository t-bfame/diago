package manager

import (
	"fmt"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	m "github.com/t-bfame/diago/internal/model"
	sto "github.com/t-bfame/diago/internal/storage"
)

type ScheduleManager interface {
	Add(schedule *m.TestSchedule, store bool) error
	Remove(id m.TestScheduleID) error
	ValidateSpec(spec string) error
}

type ScheduleManagerImpl struct {
	specParser cron.Parser
	entries    map[m.TestScheduleID]cron.EntryID
	cronRunner *cron.Cron
	jf         JobFunnel
}

func (sm *ScheduleManagerImpl) Add(schedule *m.TestSchedule, store bool) error {
	if store {
		if err := sto.AddTestSchedule(schedule); err != nil {
			return err
		}
	}

	entryID, err := sm.cronRunner.AddFunc(schedule.CronSpec, func() {
		log.WithField("TestScheduleID", schedule.ID).
			Info("About to start scheduled test")
		err := sm.jf.BeginTest(
			schedule.TestID,
			"scheduled",
		)
		if err != nil {
			log.WithField("TestScheduleID", schedule.ID).
				WithError(err).
				Errorf("Scheduled test failed to start")
		}
	})

	if err != nil {
		return err
	}

	sm.entries[schedule.ID] = entryID
	log.WithField("TestScheduleID", schedule.ID).
		Info("Added TestSchedule")

	return nil
}

func (sm *ScheduleManagerImpl) Remove(id m.TestScheduleID) error {
	entryID, exists := sm.entries[id]
	if !exists {
		return fmt.Errorf("TestSchedule<%s> is not currently running", id)
	}

	sm.cronRunner.Remove(entryID)
	delete(sm.entries, id)

	if err := sto.DeleteTestSchedule(id); err != nil {
		return err
	}

	log.WithField("TestScheduleID", id).
		Info("Removed TestSchedule")

	return nil
}

func (sm *ScheduleManagerImpl) ValidateSpec(spec string) error {
	_, err := sm.specParser.Parse(spec)
	return err
}

func (sm *ScheduleManagerImpl) onStart() {
	schedules, err := sto.GetAllTestSchedules()
	if err != nil {
		log.WithError(err).Error("ScheduleManager failed to retrieve schedules")
	}
	for _, s := range schedules {
		sm.Add(s, false)
	}
	sm.cronRunner.Start()
	log.Info("ScheduleManager started cron runner")
}

func NewScheduleManager(jf JobFunnel) ScheduleManager {
	sm := &ScheduleManagerImpl{
		// standardParser according to https://github.com/robfig/cron/blob/v3.0.1/parser.go#L217
		cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month |
				cron.Dow | cron.Descriptor,
		),
		map[m.TestScheduleID]cron.EntryID{},
		cron.New(),
		jf,
	}
	sm.onStart()
	return sm
}

type TestingScheduleManager struct {
	Added   []m.TestScheduleID
	Removed []m.TestScheduleID
}

func (sm *TestingScheduleManager) Add(
	schedule *m.TestSchedule,
	store bool,
) error {
	sm.Added = append(sm.Added, schedule.ID)
	return nil
}
func (sm *TestingScheduleManager) Remove(
	id m.TestScheduleID,
) error {
	sm.Removed = append(sm.Removed, id)
	return nil
}
func (sm *TestingScheduleManager) ValidateSpec(
	spec string,
) error {
	return nil
}
