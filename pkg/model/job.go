package model

type JobID string

type Job struct {
	ID         JobID
	Name       string
	Group      string
	Priority   int
	Env        map[string]string
	Config     []string
	Frequency  uint64
	Duration   uint64
	HTTPMethod string
	HTTPUrl    string
}
