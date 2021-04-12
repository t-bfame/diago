package internal

import (
	"os"
	"testing"

	"github.com/t-bfame/diago/pkg/model"
	sto "github.com/t-bfame/diago/pkg/storage"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testDBName string = "storageTest.db"

	jobId1          model.JobID          = "job-id-1"
	jobId2          model.JobID          = "job-id-2"
	testIdPrefix    string               = "test-id-"
	testId1         model.TestID         = "test-id-1"
	testId2         model.TestID         = "test-id-2"
	testInstanceId1 model.TestInstanceID = "testInstance-id-1"
	testInstanceId2 model.TestInstanceID = "testInstance-id-2"
	testScheduleId1 model.TestScheduleID = "testSchedule-id-1"
	testScheduleId2 model.TestScheduleID = "testSchedule-id-2"
)

var (
	job1 = &model.Job{
		ID:         jobId1,
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
	job2 = &model.Job{
		ID:         jobId2,
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
	test1 = &model.Test{
		ID:   testId1,
		Name: "test1-name",
		Jobs: []model.Job{*job1},
	}
	test2 = &model.Test{
		ID:   testId2,
		Name: "test1-name",
		Jobs: []model.Job{*job2},
	}
	testInstance1 = &model.TestInstance{
		ID:        testInstanceId1,
		TestID:    testId1,
		Type:      "typeA",
		Status:    "Pending",
		CreatedAt: 0,
		Metrics:   nil,
	}
	testInstance2 = &model.TestInstance{
		ID:        testInstanceId2,
		TestID:    testId1,
		Type:      "typeA",
		Status:    "Pending",
		CreatedAt: 0,
		Metrics:   nil,
	}
	testSchedule1 = &model.TestSchedule{
		ID:       testScheduleId1,
		Name:     "schedule-1",
		TestID:   testId1,
		CronSpec: "random-spec",
	}
	testSchedule2 = &model.TestSchedule{
		ID:       testScheduleId2,
		Name:     "schedule-2",
		TestID:   testId1,
		CronSpec: "random-spec",
	}
)

func TestAddAndGetJob(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddJob(job1); err != nil {
		t.Error("Failed to add job 1")
	}

	retrievedJob, err := sto.GetJobByJobId(jobId1)
	if err != nil {
		t.Error("Error getting job 1")
	} else {
		assert.Equal(t, job1, retrievedJob)
	}
}

func TestAddAndGetAllJobs(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddJob(job1); err != nil {
		t.Error("Failed to add job 1")
	}
	if err := sto.AddJob(job2); err != nil {
		t.Error("Failed to add job 2")
	}

	retrievedJobs, err := sto.GetAllJobs()
	if err != nil {
		t.Error("Error getting all jobs")
	} else {
		assert.ElementsMatch(t, retrievedJobs, []*model.Job{job1, job2})
	}
}

func TestAddAndDeleteJob(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddJob(job1); err != nil {
		t.Error("Failed to add job 1")
	}

	if err := sto.DeleteJob(jobId1); err != nil {
		t.Error("Failed to delete job 1")
	}

	retrievedJob, err := sto.GetJobByJobId(jobId1)
	if err != nil {
		t.Error("Error getting job 1")
	} else {
		assert.Nil(t, retrievedJob)
	}
}

func TestAddAndGetTest(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTest(test1); err != nil {
		t.Error("Failed to add test 1")
	}

	retrievedTest, err := sto.GetTestByTestId(testId1)
	if err != nil {
		t.Error("Error getting test 1")
	} else {
		assert.Equal(t, test1, retrievedTest)
	}
}

func TestAddAndGetAllTests(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTest(test1); err != nil {
		t.Error("Failed to add test 1")
	}
	if err := sto.AddTest(test2); err != nil {
		t.Error("Failed to add test 2")
	}

	retrievedTests, err := sto.GetAllTests()
	if err != nil {
		t.Error("Error getting all tests")
	} else {
		assert.ElementsMatch(t, retrievedTests, []*model.Test{test1, test2})
	}
}

func TestAddAndGetAllTestsWithPrefix(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTest(test1); err != nil {
		t.Error("Failed to add test 1")
	}
	if err := sto.AddTest(test2); err != nil {
		t.Error("Failed to add test 2")
	}

	retrievedTests, err := sto.GetAllTestsWithPrefix(testIdPrefix)
	if err != nil {
		t.Error("Error getting all tests")
	} else {
		assert.ElementsMatch(t, retrievedTests, []*model.Test{test1, test2})
	}
}

func TestGetAllTestsWithPrefixDoesNotReturn(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	retrievedTests, err := sto.GetAllTestsWithPrefix(testIdPrefix)
	if err != nil {
		t.Error("Error getting all tests")
	} else {
		assert.Empty(t, retrievedTests)
	}
}

func TestAddAndDeleteTest(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTest(test1); err != nil {
		t.Error("Failed to add test 1")
	}

	if err := sto.DeleteTest(testId1); err != nil {
		t.Error("Failed to delete test 1")
	}

	retrievedTest, err := sto.GetTestByTestId(testId1)
	if err != nil {
		t.Error("Error getting test 1")
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
		t.Error("Error getting test instance 1")
	} else {
		assert.Equal(t, testInstance1, retrievedTestInstance)
	}
}

func TestAddAndGetAllTestInstances(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestInstance(testInstance1); err != nil {
		t.Error("Failed to add test instance 1")
	}
	if err := sto.AddTestInstance(testInstance2); err != nil {
		t.Error("Failed to add test instance 2")
	}

	retrievedTestInstances, err := sto.GetAllTestInstances()
	if err != nil {
		t.Error("Error getting all test instances")
	} else {
		assert.ElementsMatch(t, retrievedTestInstances, []*model.TestInstance{testInstance1, testInstance2})
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
		t.Error("Error getting test instances")
	} else {
		assert.ElementsMatch(t, retrievedTestInstances, []*model.TestInstance{testInstance1, testInstance2})
	}
}

func TestAddAndGetTestInstancesByTestID(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestInstance(testInstance1); err != nil {
		t.Error("Failed to add test instance 1")
	}
	if err := sto.AddTestInstance(testInstance2); err != nil {
		t.Error("Failed to add test instance 2")
	}

	retrievedTestInstances, err := sto.GetTestInstancesByTestID(testId1)
	if err != nil {
		t.Error("Error getting test instances")
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

	retrievedTestInstances, err := sto.GetTestInstancesByTestID(testId1)
	if err != nil {
		t.Error("Error getting test instance 1")
	} else {
		assert.Empty(t, retrievedTestInstances)
	}
}

func TestAddAndGetTestSchedule(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestSchedule(testSchedule1); err != nil {
		t.Error("Failed to add test schedule 1")
	}

	retrievedTestSchedule, err := sto.GetTestSchedule(testScheduleId1)
	if err != nil {
		t.Error("Error getting test schedule 1")
	} else {
		assert.Equal(t, testSchedule1, retrievedTestSchedule)
	}
}

func TestAddAndGetAllTestSchedules(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestSchedule(testSchedule1); err != nil {
		t.Error("Failed to add test schedule 1")
	}
	if err := sto.AddTestSchedule(testSchedule2); err != nil {
		t.Error("Failed to add test schedule 2")
	}

	retrievedTestSchedules, err := sto.GetAllTestSchedules()
	if err != nil {
		t.Error("Error getting all test schedules")
	} else {
		assert.ElementsMatch(t, retrievedTestSchedules, []*model.TestSchedule{testSchedule1, testSchedule2})
	}
}

func TestAddAndGetTestSchedules(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestSchedule(testSchedule1); err != nil {
		t.Error("Failed to add test schedule 1")
	}
	if err := sto.AddTestSchedule(testSchedule2); err != nil {
		t.Error("Failed to add test schedule 2")
	}

	retrievedTestSchedules, err := sto.GetTestSchedules([]model.TestScheduleID{testScheduleId1, testScheduleId2})
	if err != nil {
		t.Error("Error getting test schedules")
	} else {
		assert.ElementsMatch(t, retrievedTestSchedules, []*model.TestSchedule{testSchedule1, testSchedule2})
	}
}

func TestAddAndGetTestSchedulesByTestID(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestSchedule(testSchedule1); err != nil {
		t.Error("Failed to add test schedule 1")
	}
	if err := sto.AddTestSchedule(testSchedule2); err != nil {
		t.Error("Failed to add test schedule 2")
	}

	retrievedTestSchedules, err := sto.GetTestSchedulesByTestID(testId1)
	if err != nil {
		t.Error("Error getting test schedules")
	} else {
		assert.ElementsMatch(t, retrievedTestSchedules, []*model.TestSchedule{testSchedule1, testSchedule2})
	}
}

func TestAddAndDeleteTestSchedule(t *testing.T) {
	initTestDB(t)
	defer removeTestDB()

	if err := sto.AddTestSchedule(testSchedule1); err != nil {
		t.Error("Failed to add test schedule 1")
	}

	if err := sto.DeleteTestSchedule(testScheduleId1); err != nil {
		t.Error("Failed to delete test schedule 1")
	}

	retrievedTestSchedule, err := sto.GetTestSchedule(testScheduleId1)
	if err != nil {
		t.Error("Error getting test schedule 1")
	} else {
		assert.Nil(t, retrievedTestSchedule)
	}

	retrievedTestSchedules, err := sto.GetTestSchedulesByTestID(testId1)
	if err != nil {
		t.Error("Error getting test schedules")
	} else {
		assert.Empty(t, retrievedTestSchedules)
	}
}

func initTestDB(t *testing.T) {
	if err := sto.InitDatabase(testDBName); err != nil {
		t.Error("Failed to init database")
	}
}

func removeTestDB() {
	if err := os.Remove(testDBName); err != nil {
		log.Error("Failed to remove testDB after running a test1")
	}
}
