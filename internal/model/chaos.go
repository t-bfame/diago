package model

type ChaosID string

type ChaosInstance struct {
	ID        ChaosID
	Namespace string            // Required
	Selectors map[string]string // Required
	Timeout   uint64            // Required
	Count     int               // Required
}

type ChaosResult struct {
	Status ChaosStatus
	DeletedPods []string
	Error string
}

type ChaosStatus string

const (
	ChaosFail ChaosStatus = "failed"
	ChaosSuccess ChaosStatus = "success"
)