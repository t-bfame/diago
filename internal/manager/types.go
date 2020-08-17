package manager

// JobID Unique job identifier
type JobID string

// Job configuration to be passed to scheduler
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
