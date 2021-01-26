package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
	m "github.com/t-bfame/diago/internal/model"
	sch "github.com/t-bfame/diago/internal/scheduler"
	sto "github.com/t-bfame/diago/internal/storage"
)

// APIServer serves API calls over HTTP
type APIServer struct {
	scheduler *sch.Scheduler
	jf        *mgr.JobFunnel
	sm        *mgr.ScheduleManager
}

func preResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func buildSuccess(payload interface{}, w http.ResponseWriter) []byte {
	respMap := make(map[string]interface{})
	respMap["success"] = true
	respMap["payload"] = payload

	json, err := json.Marshal(respMap)
	if err != nil {
		return buildFailure(err.Error(), http.StatusInternalServerError, w)
	}
	return json
}

func buildFailure(msg string, code int, w http.ResponseWriter) []byte {
	w.WriteHeader(code)

	respMap := make(map[string]interface{})
	respMap["success"] = false
	errMap := make(map[string]interface{})
	errMap["message"] = msg
	errMap["code"] = code
	respMap["error"] = errMap

	json, err := json.Marshal(respMap)
	if err != nil {
		log.WithField("RespMap", respMap).Info("failed to build failure")
		return make([]byte, 0)
	}
	return json
}

// Start starts the APIServer
func (server *APIServer) Start(router *mux.Router) {
	router.Use(preResponse)

	// Test C
	router.HandleFunc("/tests", func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		err = m.Validate(reflect.TypeOf(m.Test{}), bodyContent)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		var test m.Test
		err = json.Unmarshal(bodyContent, &test)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		// For now, make id by using random hash of length 5
		// TODO: Maybe use a counter for every group for better UX?
		testid := test.Name
		test.ID = m.TestID(testid)

		for i := range test.Jobs {
			test.Jobs[i].ID = m.JobID(fmt.Sprintf("%s-%d", test.ID, i))
		}

		err = sto.AddTest(&test)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusInternalServerError, w))
			return
		}

		w.Write(
			buildSuccess(
				map[string]string{
					"testid": testid,
				},
				w,
			),
		)

		log.WithField("TestID", testid).Info("Test created")

	}).Methods(http.MethodPost)

	// Test RUD
	router.HandleFunc("/tests/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]

		// make sure Test exists
		test, err := sto.GetTestByTestId(m.TestID(testid))
		if err != nil {
			w.Write(buildFailure(
				fmt.Sprintf("Cannot find Test<%s>", testid),
				http.StatusNotFound,
				w,
			))
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Write(buildSuccess(test, w))

			log.WithField("TestID", testid).Info("Test retrieved")
		case http.MethodPut:
			bodyContent, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
				return
			}

			err = m.Validate(reflect.TypeOf(m.Test{}), bodyContent)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
				return
			}

			var updatedTest m.Test
			err = json.Unmarshal(bodyContent, &updatedTest)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
				return
			}
			updatedTest.ID = m.TestID(testid)
			for i := range updatedTest.Jobs {
				updatedTest.Jobs[i].ID = m.JobID(fmt.Sprintf("%s-%d", updatedTest.ID, i))
			}
			sto.AddTest(&updatedTest)
			w.Write(buildSuccess(updatedTest, w))

			log.WithField("TestID", testid).Info("Test updated")
		case http.MethodDelete:
			w.Write(
				buildFailure(
					"Deletion not implemented",
					http.StatusNotImplemented,
					w,
				),
			)
		default:
			w.Write(buildFailure("Request not supported", http.StatusBadRequest, w))
		}
	}).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)

	// Get TestInstances by TestID
	router.HandleFunc("/test-instances/{testid}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["testid"]
		instances, err := sto.GetTestInstancesByTestID(m.TestID(testid))
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusNotFound, w))
			return
		}
		w.Write(buildSuccess(instances, w))

		log.WithField("TestID", testid).Info("Test instances retrieved")
	}).Methods(http.MethodGet)

	// Start adhoc
	router.HandleFunc("/start-test/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]
		err := server.jf.BeginTest(m.TestID(testid), "adhoc")
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		w.Write(
			buildSuccess(
				fmt.Sprintf("Successfully submitted Test<%s>", testid),
				w,
			),
		)
	})

	// stop TestInstance by TestID
	router.HandleFunc("/stop-test/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]
		err := server.jf.StopTest(m.TestID(testid))
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		w.Write(
			buildSuccess(
				fmt.Sprintf("Successfully stopped Test<%s>", testid),
				w,
			),
		)
	})

	prepareTestSchedule := func(w http.ResponseWriter, r *http.Request) *m.TestSchedule {
		bodyContent, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return nil
		}

		err = m.Validate(reflect.TypeOf(m.TestSchedule{}), bodyContent)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return nil
		}

		var schedule m.TestSchedule
		err = json.Unmarshal(bodyContent, &schedule)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return nil
		}

		// make sure cron spec is valid
		err = server.sm.ValidateSpec(schedule.CronSpec)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return nil
		}

		// make sure specified Test exists
		if _, err := sto.GetTestByTestId(schedule.TestID); err != nil {
			w.Write(buildFailure(
				fmt.Sprintf("Cannot find Test<%s>", schedule.TestID),
				http.StatusNotFound,
				w,
			))
			return nil
		}

		schedule.ID = m.TestScheduleID(schedule.Name)

		return &schedule
	}

	router.HandleFunc("/test-schedules", func(w http.ResponseWriter, r *http.Request) {
		schedule := prepareTestSchedule(w, r)
		if schedule == nil {
			return
		}

		server.sm.Add(schedule)

		w.Write(
			buildSuccess(
				map[string]string{
					"scheduleid": string(schedule.ID),
				},
				w,
			),
		)
	}).Methods(http.MethodPost)

	router.HandleFunc("/test-schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		scheduleid := vars["id"]

		switch r.Method {
		case http.MethodGet:
			// TODO: get from storage
			w.Write(
				buildSuccess(
					map[string]interface{}{},
					w,
				),
			)
		case http.MethodPut:
			// TODO: check if TestSchedule actually exists

			schedule := prepareTestSchedule(w, r)
			if schedule == nil {
				return
			}

			// make sure id isn't changed
			schedule.ID = m.TestScheduleID(scheduleid)

			server.sm.Update(schedule)

			w.Write(
				buildSuccess(
					map[string]string{
						"scheduleid": string(schedule.ID),
					},
					w,
				),
			)
		case http.MethodDelete:
			server.sm.Remove(m.TestScheduleID(scheduleid))

			w.Write(
				buildSuccess(
					map[string]string{
						"scheduleid": scheduleid,
					},
					w,
				),
			)
		}
	}).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
}

// NewAPIServer create a new APIServer
func NewAPIServer(sched *sch.Scheduler, jf *mgr.JobFunnel, sm *mgr.ScheduleManager) *APIServer {
	return &APIServer{sched, jf, sm}
}
