package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/t-bfame/diago/pkg/utils"
)

type (
	Test   map[string]interface{}
	TestID string
)

type (
	Job   map[string]interface{}
	JobID string
)

var (
	testValidator      validator
	jobValidator       validator
	transformValidator validator
)

func Test_Collection_Name() string {
	return "test"
}

func Test_Collection() map[string]Test {
	return *(storage[Test_Collection_Name()].(*map[string]Test))
}

func InitTest() {
	testCollection := make(map[string]Test)
	storage[Test_Collection_Name()] = &(testCollection)

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
	}, Test_Collection_Name())

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

func (test Test) save() (bool, error) {
	if ok, et := testValidator(map[string]interface{}(test)); !ok {
		return false, errors.New(et.String())
	}

	Test_Collection()[string(test["ID"].(TestID))] = test
	return true, nil
}

func TestByID(id TestID) Test {
	return Test_Collection()[string(id)]
}

func (test Test) delete() (bool, error) {
	key := test["ID"].(string)
	_, ok := Test_Collection()[key]
	if ok {
		delete(Test_Collection(), key)
		return true, nil
	}
	return false, errors.New("Test not found")
}

func CreateTest(content map[string]interface{}) (Test, error) {
	// Validate unmarshalled json first, so we know what to expect
	if ok, et := testValidator(content); !ok {
		return nil, errors.New(et.String())
	}

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

	test := Test(content)
	if ok, err := test.save(); !ok {
		return nil, err
	}
	return test, nil
}

func TestFromBody(body []byte) (map[string]interface{}, error) {
	content := make(map[string]interface{})
	d := json.NewDecoder(bytes.NewBuffer(body))
	d.UseNumber()
	if err := d.Decode(&content); err != nil {
		return nil, err
	}

	// Pre-verify fields to transform
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

	return content, nil
}
