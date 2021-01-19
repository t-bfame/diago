package internal

import (
	"os"
	"testing"

	"github.com/t-bfame/diago/internal/model"
	sto "github.com/t-bfame/diago/internal/storage"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testDBName string = "storageTest.db"

	jobId           model.JobID          = "job-id"
	testId          model.TestID         = "test-id"
	testInstanceId1 model.TestInstanceID = "testInstance-id-1"
	testInstanceId2 model.TestInstanceID = "testInstance-id-2"
)

var (
	job = &model.Job{
		ID:         jobId,
		Name:       "name",
		Group:      "group",
		Priority:   1,
		Env:        map[string]string{"1": "1"},
		Config:     []string{"config"},
		Frequency:  10,
		Duration:   100,
		HTTPMethod: "GET",
		HTTPUrl:    "localhost",
	}
	test = &model.Test{
		ID:   testId,
		Name: "test-name",
		Jobs: []model.Job{*job},
	}
	testInstance1 = &model.TestInstance{
		ID:        testInstanceId1,
		TestID:    testId,
		Type:      "typeA",
		Status:    "Pending",
		CreatedAt: 0,
		Metrics:   nil,
	}
	testInstance2 = &model.TestInstance{
		ID:        testInstanceId2,
		TestID:    testId,
		Type:      "typeA",
		Status:    "Pending",
		CreatedAt: 0,
		Metrics:   nil,
	}
)

func TestAddAndGetJob(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddJob(job); err != nil {
		t.Error("Failed to add job")
	}

	retrievedJob, err := sto.GetJobByJobId(jobId)
	if err != nil {
		t.Error("Failed to get job")
	} else {
		assert.Equal(t, job, retrievedJob)
	}
}

func TestAddAndDeleteJob(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddJob(job); err != nil {
		t.Error("Failed to add job")
	}

	if err := sto.DeleteJob(jobId); err != nil {
		t.Error("Failed to delete job")
	}

	retrievedJob, err := sto.GetJobByJobId(jobId)
	if err != nil {
		t.Error("Failed to get job")
	} else {
		assert.Nil(t, retrievedJob)
	}
}

func TestAddAndGetTest(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTest(test); err != nil {
		t.Error("Failed to add test")
	}

	retrievedTest, err := sto.GetTestByTestId(testId)
	if err != nil {
		t.Error("Failed to get test")
	} else {
		assert.Equal(t, test, retrievedTest)
	}
}

func TestAddAndDeleteTest(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTest(test); err != nil {
		t.Error("Failed to add test")
	}

	if err := sto.DeleteTest(testId); err != nil {
		t.Error("Failed to delete test")
	}

	retrievedTest, err := sto.GetTestByTestId(testId)
	if err != nil {
		t.Error("Failed to get test")
	} else {
		assert.Nil(t, retrievedTest)
	}
}

func TestAddAndGetTestInstance(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestInstance(testInstance1); err != nil {
		t.Error("Failed to add test instance 1")
	}

	retrievedTestInstance, err := sto.GetTestInstance(testInstanceId1)
	if err != nil {
		t.Error("Failed to get test instance 1")
	} else {
		assert.Equal(t, testInstance1, retrievedTestInstance)
	}
}

func TestAddAndGetTestInstances(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestInstance(testInstance1); err != nil {
		t.Error("Failed to add test instance 1")
	}
	if err := sto.AddTestInstance(testInstance2); err != nil {
		t.Error("Failed to add test instance 2")
	}

	retrievedTestInstances, err := sto.GetTestInstances([]model.TestInstanceID{testInstanceId1, testInstanceId2})
	if err != nil {
		t.Error("Failed to get test instance")
	} else {
		assert.ElementsMatch(t, retrievedTestInstances, []*model.TestInstance{testInstance1, testInstance2})
	}
}

func TestAddAndDeleteTestInstance(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestInstance(testInstance1); err != nil {
		t.Error("Failed to add test instance 1")
	}

	if err := sto.DeleteTestInstance(testInstanceId1); err != nil {
		t.Error("Failed to delete test instance 1")
	}

	retrievedTestInstance, err := sto.GetTestInstance(testInstanceId1)
	if err != nil {
		t.Error("Failed to get test instance 1")
	} else {
		assert.Nil(t, retrievedTestInstance)
	}
}

func initTestDB(t *testing.T) {
	if err := sto.InitDatabase(testDBName); err != nil {
		t.Error("Failed to init database")
	}
}

func removeTestDB() {
	if err := os.Remove(testDBName); err != nil {
		log.Error("Failed to remove testDB after running a test")
	}
}
