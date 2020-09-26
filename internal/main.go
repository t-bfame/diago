package main

import (
	"fmt"
	"github.com/t-bfame/diago/internal/manager"
	sto "github.com/t-bfame/diago/internal/storage"
)

func main() {
	sto.InitDatabase(true)

	// region define-test-data
	var (
		jobId          manager.JobID          = "job-id"
		testId         manager.TestID         = "test-id"
		testInstanceId manager.TestInstanceID = "testInstance-id"
	)
	job := &manager.Job{
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
	test := &manager.Test{
		ID:   testId,
		Name: "test-name",
		Jobs: []manager.Job{*job},
	}
	testInstance := &manager.TestInstance{
		ID:        testInstanceId,
		TestID:    testId,
		Type:      "typeA",
		Status:    "Pending",
		CreatedAt: 0,
		Metrics:   nil,
	}
	// endregion

	// region add-objects
	_ = sto.AddJob(job)
	_ = sto.AddTest(test)
	_ = sto.AddTestInstance(testInstance)
	// endregion

	// region retrieve-objects
	retrievedJob, err := sto.GetJobByJobId(jobId)
	if err != nil {
		fmt.Printf("no")
	} else {
		fmt.Printf("yes: %+v\n", retrievedJob)
	}

	retrievedTest, err := sto.GetTestByTestId(testId)
	if err != nil {
		fmt.Printf("no")
	} else {
		fmt.Printf("yes: %+v\n", retrievedTest)
	}

	retrievedTestInstance, err := sto.GetTestInstanceByTestInstanceId(testInstanceId)
	if err != nil {
		fmt.Printf("no")
	} else {
		fmt.Printf("yes: %+v\n", retrievedTestInstance)
	}
	// endregion
}
