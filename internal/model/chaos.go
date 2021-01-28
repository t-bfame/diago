package model

type ChaosInstance struct {
	Namespace string            // Required
	Selectors map[string]string // Required
	Timeout   uint64            // Required
	Duration  uint64            // Required
	Count     int               // Optional
}
