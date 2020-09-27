package model

import (
	"fmt"
	"strconv"
	"time"
)

type TestInstanceID string

type TestInstance struct {
	ID        TestInstanceID
	TestID    TestID
	Type      string
	Status    string
	CreatedAt int64
	Metrics   interface{} // mapping of jobID -> job metrics
}

func TestInstanceCollection() map[string][]TestInstance {
	return *(storage["testinstance"].(*map[string][]TestInstance))
}

func InitTestInstance() {
	testInstanceCollection := make(map[string][]TestInstance)
	storage["testinstance"] = &(testInstanceCollection)
}

func (instance *TestInstance) Save() (*TestInstance, error) {
	key := string(instance.TestID)

	// update?
	for i, ti := range TestInstanceCollection()[key] {
		if ti.ID == instance.ID {
			TestInstanceCollection()[key][i] = *instance
			return instance, nil
		}
	}

	// create
	TestInstanceCollection()[key] =
		append(TestInstanceCollection()[key], *instance)
	return instance, nil
}

func (instance *TestInstance) IsTerminal() bool {
	return instance.Status == "failed" || instance.Status == "done"
}

func TestInstancesByTestID(id TestID) ([]TestInstance, bool) {
	instances, exists := TestInstanceCollection()[string(id)]
	return instances, exists
}

func CreateTestInstance(id TestID) (*TestInstance, error) {
	key := string(id)
	test, ok := TestCollection()[key]
	if !ok {
		return nil, fmt.Errorf("Test<%s> not found", key)
	}

	// for now generate uid using test name + timestamp
	now := time.Now().Unix()
	instanceid := test.Name + "-" + strconv.FormatInt(now, 10)

	instance := &TestInstance{
		ID:        TestInstanceID(instanceid),
		TestID:    id,
		Type:      "adhoc",
		Status:    "pending",
		CreatedAt: now,
	}
	_, err := instance.Save()
	if err != nil {
		return nil, err
	}

	return instance, nil
}
