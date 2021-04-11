package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	mgr "github.com/t-bfame/diago/internal/manager"
	m "github.com/t-bfame/diago/internal/model"
	sto "github.com/t-bfame/diago/internal/storage"
)

const (
	uri        = "testing.diago.com/api"
	testDBName = "apiserverTest.db"
)

type TestResponseWriter struct {
	headers         http.Header
	responseContent *[]byte
	status          *int
}

func (t TestResponseWriter) Header() http.Header {
	return t.headers
}

func (t TestResponseWriter) Write(content []byte) (int, error) {
	*t.responseContent = content
	return len(content), nil
}

func (t TestResponseWriter) WriteHeader(statusCode int) {
	*t.status = statusCode
}

func TestHandleTestCreate(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	invalid := []byte(`{"Unexpected": 1}`)
	r, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(invalid))
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestCreate(w, r)
	if status != http.StatusBadRequest {
		t.Errorf("Expected TestCreate to fail")
	}

	test := []byte(
		`{
			"Name": "Test1",
			"Jobs": [
				{
					"Name": "Job1",
					"Group": "test-worker",
					"Priority": 0,
					"Frequency": 5,
					"Duration": 30,
					"HTTPMethod": "GET",
					"HTTPUrl": "https://www.google.com"
				}
			]
		}`,
	)
	r, _ = http.NewRequest(http.MethodPost, uri, bytes.NewReader(test))
	content, status = []byte(``), http.StatusOK
	w = TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestCreate(w, r)
	if status != http.StatusOK {
		t.Errorf("Expected TestCreate to pass")
	}
	tm, err := sto.GetTestByTestId(m.TestID("Test1"))
	if err != nil || tm == nil {
		t.Errorf("Expected TestCreate to persist")
	}
}

func TestHandleTestReadList(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	test1 := m.Test{
		ID:   "ABC",
		Name: "ABC",
		Jobs: []m.Job{},
	}
	test2 := m.Test{
		ID:   "ABD",
		Name: "ABD",
		Jobs: []m.Job{},
	}
	test3 := m.Test{
		ID:   "B",
		Name: "B",
		Jobs: []m.Job{},
	}

	sto.AddTest(&test1)
	sto.AddTest(&test2)
	sto.AddTest(&test3)

	// ReadForPrefix
	r, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s?prefix=A", uri),
		bytes.NewReader([]byte(``)),
	)
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestReadForPrefix(w, r)
	var result interface{}
	json.Unmarshal(content, &result)
	tests := result.(map[string]interface{})["payload"].([]interface{})
	for _, tst := range tests {
		name := (tst.(map[string]interface{}))["Name"].(string)
		if !strings.HasPrefix(name, "A") {
			t.Errorf("Test `%s` does not have prefix `A`", name)
		}
	}
	if len(tests) != 2 {
		t.Errorf("Expected 2 results for TestReadForPrefix, got %d", len(tests))
	}

	// ReadAll
	r, _ = http.NewRequest(http.MethodGet, uri, bytes.NewReader([]byte(``)))
	content, status = []byte(``), http.StatusOK
	w = TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestReadAll(w, r)
	json.Unmarshal(content, &result)
	tests = result.(map[string]interface{})["payload"].([]interface{})
	if len(tests) != 3 {
		t.Errorf("Expected 3 results for TestReadAll, got %d", len(tests))
	}

}

func TestHandleTestRead(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	test := m.Test{
		ID:   "B",
		Name: "B",
		Jobs: []m.Job{},
	}
	sto.AddTest(&test)

	// Read 404
	r, _ := http.NewRequest(
		http.MethodGet,
		uri,
		bytes.NewReader([]byte(``)),
	)
	r = mux.SetURLVars(r, map[string]string{
		"testid": "TestNF",
	})
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestRead(w, r)
	if status != http.StatusNotFound {
		t.Errorf("Expected TestRead to fail")
	}

	// Read
	r, _ = http.NewRequest(
		http.MethodGet,
		uri,
		bytes.NewReader([]byte(``)),
	)
	r = mux.SetURLVars(r, map[string]string{
		"testid": "B",
	})
	content, status = []byte(``), http.StatusOK
	w = TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestRead(w, r)
	if status != http.StatusOK {
		t.Errorf("Expected TestRead to succeed")
	}
	var result interface{}
	json.Unmarshal(content, &result)
	job := result.(map[string]interface{})["payload"]
	name := job.(map[string]interface{})["Name"]
	if name != "B" {
		t.Errorf("Expected TestRead to retrieve `B`, got `%s`", name)
	}
}

func TestHandleDelete(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	test := m.Test{
		ID:   "B",
		Name: "B",
		Jobs: []m.Job{},
	}
	sto.AddTest(&test)

	r, _ := http.NewRequest(
		http.MethodDelete,
		uri,
		bytes.NewReader([]byte(``)),
	)
	r = mux.SetURLVars(r, map[string]string{
		"testid": "B",
	})
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestDelete(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestDelete to succeed")
	}

	tm, _ := sto.GetTestByTestId("B")
	if tm != nil {
		t.Error("Expected TestDelete to persist")
	}
}

func TestHandleTestStart(t *testing.T) {
	jf := &mgr.TestingJobFunnel{}
	sm := &mgr.TestingScheduleManager{}
	server := NewAPIServer(jf, sm)
	startHandler := handleTestStartBuilder(server)

	r, _ := http.NewRequest(
		http.MethodPost,
		uri,
		bytes.NewReader([]byte(``)),
	)
	r = mux.SetURLVars(r, map[string]string{
		"testid": "Test1",
	})
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	startHandler(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestStart to succeed")
	}
	if len(jf.Starts) != 1 || string(jf.Starts[0]) != "Test1" {
		t.Error("Expected TestStart to start test")
	}
}

func TestHandleTestStop(t *testing.T) {
	jf := &mgr.TestingJobFunnel{}
	sm := &mgr.TestingScheduleManager{}
	server := NewAPIServer(jf, sm)
	stopHandler := handleTestStopBuilder(server)

	r, _ := http.NewRequest(
		http.MethodPost,
		uri,
		bytes.NewReader([]byte(``)),
	)
	r = mux.SetURLVars(r, map[string]string{
		"testid": "Test1",
	})
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	stopHandler(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestStop to succeed")
	}
	if len(jf.Stops) != 1 || string(jf.Stops[0]) != "Test1" {
		t.Error("Expected TestStop to stop test")
	}
}

func TestHandleTestInstanceRead(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	ti1 := m.TestInstance{
		ID:        "Test1-1618100600",
		TestID:    "Test1",
		Type:      "adhoc",
		Status:    "submitted",
		CreatedAt: 1618100600,
	}
	ti2 := m.TestInstance{
		ID:        "Test2-1618100600",
		TestID:    "Test2",
		Type:      "adhoc",
		Status:    "submitted",
		CreatedAt: 1618100600,
	}

	sto.AddTestInstance(&ti1)
	sto.AddTestInstance(&ti2)

	r, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s?testid=Test1", uri),
		bytes.NewReader([]byte(``)),
	)
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestInstanceReadForTest(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestInstanceReadForTest to succeed")
	}
	var result interface{}
	json.Unmarshal(content, &result)
	tis := result.(map[string]interface{})["payload"].([]interface{})
	if len(tis) != 1 {
		t.Errorf("Expected 1 result for TestInstanceReadForTest, got %d", len(tis))
	}
	tiID := tis[0].(map[string]interface{})["ID"]
	if tiID != "Test1-1618100600" {
		t.Errorf(
			"Expected TestInstanceReadForTest to retrieve `Test1-1618100600`, got `%s`",
			tiID,
		)
	}

	r, _ = http.NewRequest(
		http.MethodGet,
		uri,
		bytes.NewReader([]byte(``)),
	)
	content, status = []byte(``), http.StatusOK
	w = TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestInstanceReadAll(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestInstanceReadAll to succeed")
	}
	json.Unmarshal(content, &result)
	tis = result.(map[string]interface{})["payload"].([]interface{})
	tiCount := len(tis)
	if tiCount != 2 {
		t.Errorf("Expected 2 results for TestInstanceReadAll, got %d", tiCount)
	}
}

func TestHandleTestScheduleCreate(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	jf := &mgr.TestingJobFunnel{}
	sm := &mgr.TestingScheduleManager{}
	server := NewAPIServer(jf, sm)
	tsCreateHandler := handleTestScheduleCreateBuilder(server)

	ts := []byte(
		`{
			"Name": "TestSchedule1",
			"TestID": "Test1",
			"CronSpec": "* * * * *"
		}`,
	)

	r, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(ts))
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	tsCreateHandler(w, r)
	if status != http.StatusBadRequest {
		t.Error("Expected TestScheduleCreate to fail")
	}

	test := m.Test{
		ID:   "Test1",
		Name: "Test1",
		Jobs: []m.Job{},
	}
	sto.AddTest(&test)

	r, _ = http.NewRequest(http.MethodPost, uri, bytes.NewReader(ts))
	content, status = []byte(``), http.StatusOK
	w = TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	tsCreateHandler(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestScheduleCreate to succeed")
	}
	if len(sm.Added) != 1 || string(sm.Added[0]) != "TestSchedule1" {
		t.Error("Expected TestScheduleCreate to add schedule")
	}
}

func TestHandleTestScheduleRead(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	ts1 := m.TestSchedule{
		ID:       "TestSchedule1",
		Name:     "TestSchedule1",
		TestID:   "Test1",
		CronSpec: "* * * * *",
	}
	ts2 := m.TestSchedule{
		ID:       "TestSchedule2",
		Name:     "TestSchedule2",
		TestID:   "Test2",
		CronSpec: "* * * * *",
	}
	sto.AddTestSchedule(&ts1)
	sto.AddTestSchedule(&ts2)

	r, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s?testid=Test1", uri),
		bytes.NewReader([]byte(``)),
	)
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestScheduleReadForTest(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestScheduleReadForTest to succeed")
	}
	var result interface{}
	json.Unmarshal(content, &result)
	schedules := result.(map[string]interface{})["payload"].([]interface{})
	if len(schedules) != 1 {
		t.Errorf("Expected 1 result for TestScheduleReadForTest, got %d", len(schedules))
	}
	sid := schedules[0].(map[string]interface{})["ID"]
	if sid != "TestSchedule1" {
		t.Errorf("Expected TestScheduleReadForTest to retrieve `TestSchedule1`, got `%s`", sid)
	}

	r, _ = http.NewRequest(
		http.MethodGet,
		uri,
		bytes.NewReader([]byte(``)),
	)
	content, status = []byte(``), http.StatusOK
	w = TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	handleTestScheduleReadAll(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestScheduleReadAll to succeed")
	}
	json.Unmarshal(content, &result)
	schedules = result.(map[string]interface{})["payload"].([]interface{})
	if len(schedules) != 2 {
		t.Errorf("Expected 2 results for TestScheduleReadAll, got %d", len(schedules))
	}
}

func TestHandleTestScheduleDelete(t *testing.T) {
	initTestDB(t)
	defer removeTestDB(t)

	jf := &mgr.TestingJobFunnel{}
	sm := &mgr.TestingScheduleManager{}
	server := NewAPIServer(jf, sm)
	tsDeleteHandler := handleTestScheduleDeleteBuilder(server)

	ts1 := m.TestSchedule{
		ID:       "TestSchedule1",
		Name:     "TestSchedule1",
		TestID:   "Test1",
		CronSpec: "* * * * *",
	}
	sto.AddTestSchedule(&ts1)

	r, _ := http.NewRequest(
		http.MethodDelete,
		uri,
		bytes.NewReader([]byte(``)),
	)
	r = mux.SetURLVars(r, map[string]string{
		"scheduleid": "TestSchedule1",
	})
	content, status := []byte(``), http.StatusOK
	w := TestResponseWriter{
		http.Header{},
		&content,
		&status,
	}

	tsDeleteHandler(w, r)
	if status != http.StatusOK {
		t.Error("Expected TestScheduleDelete to succeed")
	}
	if len(sm.Removed) != 1 || string(sm.Removed[0]) != "TestSchedule1" {
		t.Error("Expected TestScheduleDelete to remove schedule")
	}
}

func initTestDB(t *testing.T) {
	if err := sto.InitDatabase(testDBName); err != nil {
		t.Error("Failed to init database")
	}
}

func removeTestDB(t *testing.T) {
	if err := os.Remove(testDBName); err != nil {
		t.Log("Failed to remove testDB after running a test")
	}
}
