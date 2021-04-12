package model

type TestScheduleID string

type TestSchedule struct {
	ID       TestScheduleID
	Name     string `validation:"required"`
	TestID   TestID `validation:"required"`
	CronSpec string `validation:"required"`
}
