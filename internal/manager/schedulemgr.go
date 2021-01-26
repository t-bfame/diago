package manager

import (
	"fmt"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	m "github.com/t-bfame/diago/internal/model"
)

type ScheduleManager struct {
	specParser cron.Parser
	entries    map[m.TestScheduleID]cron.EntryID
	cronRunner *cron.Cron
	jf         *JobFunnel
}

func (sm *ScheduleManager) Add(schedule *m.TestSchedule) error {
	// TODO: add to storage

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

func (sm *ScheduleManager) Remove(id m.TestScheduleID) error {
	entryID, exists := sm.entries[id]
	if !exists {
		return fmt.Errorf("TestSchedule<%s> is not currently running", id)
	}

	sm.cronRunner.Remove(entryID)
	delete(sm.entries, id)

	log.WithField("TestScheduleID", id).
		Info("Removed TestSchedule")

	// TODO: delete from storage

	return nil
}

func (sm *ScheduleManager) Update(schedule *m.TestSchedule) error {
	err := sm.Remove(schedule.ID)
	if err != nil {
		return err
	}
	err = sm.Add(schedule)
	if err != nil {
		return err
	}
	return nil
}

func (sm *ScheduleManager) ValidateSpec(spec string) error {
	_, err := sm.specParser.Parse(spec)
	return err
}

func (sm *ScheduleManager) onStart() {
	// TODO: query storage for all TestSchedules and run all
	sm.cronRunner.Start()
	log.Info("ScheduleManager started cron runner")
}

func NewScheduleManager(jf *JobFunnel) *ScheduleManager {
	sm := &ScheduleManager{
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
