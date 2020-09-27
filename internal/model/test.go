package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/t-bfame/diago/pkg/utils"
)

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

type TestID string

type Test struct {
	ID   TestID
	Name string
	Jobs []Job
}

var (
	testValidator      validator
	jobValidator       validator
	transformValidator validator
)

func TestCollection() map[string]Test {
	return *(storage["test"].(*map[string]Test))
}

func InitTest() {
	testCollection := make(map[string]Test)
	storage["test"] = &(testCollection)

	jobValidator = doc(map[string]validator{
		"ID":         opt(typ(JobID(""))),
		"Name":       kind(reflect.String),
		"Group":      kind(reflect.String),
		"Priority":   kind(reflect.Int),
		"Env":        opt(kind(reflect.Map)),
		"Config":     opt(list(kind(reflect.String))),
		"Frequency":  kind(reflect.Uint64),
		"Duration":   kind(reflect.Uint64),
		"HTTPMethod": kind(reflect.String),
		"HTTPUrl":    kind(reflect.String),
	}, "test")

	testValidator = doc(map[string]validator{
		"ID":   opt(typ(TestID(""))),
		"Name": kind(reflect.String),
		"Jobs": list(jobValidator),
	})

	transformValidator = doc(map[string]validator{
		"Jobs": list(
			doc(map[string]validator{
				"Priority":  typ(json.Number("")),
				"Frequency": typ(json.Number("")),
				"Duration":  typ(json.Number("")),
			}),
		),
	})
}

func (test *Test) Save() (*Test, error) {
	TestCollection()[string(test.ID)] = *test
	return test, nil
}

func TestByID(id TestID) (Test, bool) {
	test, exists := TestCollection()[string(id)]
	return test, exists
}

func (test *Test) Delete() (bool, error) {
	key := string(test.ID)
	_, ok := TestCollection()[key]
	if ok {
		delete(TestCollection(), key)
		return true, nil
	}
	return false, fmt.Errorf("Test<%s> not found", key)
}

func SaveTestFromBody(body []byte, create ...bool) (*Test, error) {
	content := make(map[string]interface{})
	d := json.NewDecoder(bytes.NewBuffer(body))
	d.UseNumber()
	if err := d.Decode(&content); err != nil {
		return nil, err
	}

	// Validate fields to transform
	if ok, et := transformValidator(content); !ok {
		return nil, errors.New(et.String())
	}

	// Do transformations
	for i := range content["Jobs"].([]interface{}) {
		j := &content["Jobs"].([]interface{})[i]
		priority, _ := strconv.Atoi((*j).(map[string]interface{})["Priority"].(json.Number).String())
		frequency, _ := strconv.ParseUint((*j).(map[string]interface{})["Frequency"].(json.Number).String(), 10, 64)
		duration, _ := strconv.ParseUint((*j).(map[string]interface{})["Duration"].(json.Number).String(), 10, 64)

		(*j).(map[string]interface{})["Priority"] = priority
		(*j).(map[string]interface{})["Frequency"] = frequency
		(*j).(map[string]interface{})["Duration"] = duration
	}

	// Validate everything else
	if ok, et := testValidator(content); !ok {
		return nil, errors.New(et.String())
	}

	// Test Creation (as opposed to Update)
	if len(create) > 0 && create[0] {
		// Generate job ids
		name := content["Name"]

		// For now, make id by using random hash of length 5
		// TODO: Maybe use a counter for every group for better UX?
		testID := fmt.Sprintf("%s-%s", name, utils.RandHash(5))
		content["ID"] = TestID(testID)

		// Assign job ids
		for i := range content["Jobs"].([]interface{}) {
			j := &content["Jobs"].([]interface{})[i]
			(*j).(map[string]interface{})["ID"] = JobID(fmt.Sprintf("%s-%d", testID, i))
		}
	}

	var test Test
	mapstructure.Decode(content, &test)

	_, err := test.Save()
	if err != nil {
		return nil, err
	}

	return &test, nil
}
