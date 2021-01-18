package model

type ChaosInstance struct {
	Namespace    string
	Count        uint32
	Selectors    map[string]string
	Timeout      uint64
	TestDuration uint64
}
