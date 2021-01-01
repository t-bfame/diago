package model

type TestID string

type Test struct {
	ID   TestID
	Name string
	Jobs []Job
}
