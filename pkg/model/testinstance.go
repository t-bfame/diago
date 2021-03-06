package model

type TestInstanceID string

type TestInstance struct {
	ID        TestInstanceID
	TestID    TestID
	Type      string
	Status    string
	CreatedAt int64
	Metrics   interface{} // TODO: decide how to store metrics long-term
	ChaosResult map[ChaosID]ChaosResult
	Error string
}

func (instance *TestInstance) IsTerminal() bool {
	return instance.Status == "failed" || instance.Status == "done" || instance.Status == "stopped"
}
