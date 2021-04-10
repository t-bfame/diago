package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
	m "github.com/t-bfame/diago/internal/model"
	sch "github.com/t-bfame/diago/internal/scheduler"
	sto "github.com/t-bfame/diago/internal/storage"
	dash "github.com/t-bfame/diago/internal/dashboard"
)

// APIServer serves API calls over HTTP
type APIServer struct {
	scheduler *sch.Scheduler
	jf        *mgr.JobFunnel
	sm        *mgr.ScheduleManager
	db	      *dash.Dashboard
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

		testid := test.Name
		test.ID = m.TestID(testid)

		for i := range test.Jobs {
			test.Jobs[i].ID = m.JobID(fmt.Sprintf("%s-%d", test.ID, i))
		}

		for i := range test.Chaos {
			test.Chaos[i].ID = m.ChaosID(fmt.Sprintf("%s-%d", test.ID, i))
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

	// Test RD
	router.HandleFunc("/tests/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]

		// Handle prefix search
		// TODO: prevent "*" from being a valid character in TestIDs
		if strings.HasSuffix(testid, "*") {
			tests, err := sto.GetAllTestsWithPrefix(testid[:len(testid)-1])
			if err != nil {
				w.Write(
					buildFailure(err.Error(), http.StatusInternalServerError, w),
				)
				return
			}
			w.Write(buildSuccess(tests, w))
			return
		}

		// make sure Test exists
		test, err := sto.GetTestByTestId(m.TestID(testid))
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusInternalServerError, w))
			return
		} else if test == nil {
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
		case http.MethodDelete:
			if err := sto.DeleteTest(m.TestID(testid)); err != nil {
				w.Write(
					buildFailure(err.Error(), http.StatusInternalServerError, w),
				)
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
		default:
			w.Write(buildFailure("Request not supported", http.StatusBadRequest, w))
		}
	}).Methods(http.MethodGet, http.MethodDelete)

	// Get TestInstances by TestID
	router.HandleFunc("/test-instances/{testid}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["testid"]

		// TODO: prevent "all" from being a useable TestID
		if testid == "all" {
			tis, err := sto.GetAllTestInstances()
			if err != nil {
				w.Write(
					buildFailure(err.Error(), http.StatusInternalServerError, w),
				)
				return
			}
			w.Write(buildSuccess(tis, w))
			return
		}

		instances, err := sto.GetTestInstancesByTestID(m.TestID(testid))
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusInternalServerError, w))
			return
		} else if instances == nil {
			// empty
			instances = []*m.TestInstance{}
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
		test, err := sto.GetTestByTestId(schedule.TestID)
		if err != nil {
			w.Write(
				buildFailure(err.Error(), http.StatusInternalServerError, w),
			)
			return nil
		} else if test == nil {
			w.Write(buildFailure(
				fmt.Sprintf("Cannot find Test<%s>", schedule.TestID),
				http.StatusBadRequest,
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

		if err := server.sm.Add(schedule, true); err != nil {
			w.Write(
				buildFailure(err.Error(), http.StatusInternalServerError, w),
			)
			return
		}

		w.Write(
			buildSuccess(
				map[string]string{
					"scheduleid": string(schedule.ID),
				},
				w,
			),
		)
	}).Methods(http.MethodPost)

	// Get TestSchedules by TestID
	router.HandleFunc("/test-schedules/{testid}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["testid"]

		// TODO: prevent "all" from being a useable TestID
		if testid == "all" {
			schedules, err := sto.GetAllTestSchedules()
			if err != nil {
				w.Write(
					buildFailure(err.Error(), http.StatusInternalServerError, w),
				)
				return
			}
			w.Write(buildSuccess(schedules, w))
			return
		}

		// schedules exist?
		schedules, err := sto.GetTestSchedulesByTestID(m.TestID(testid))
		if err != nil {
			w.Write(
				buildFailure(err.Error(), http.StatusInternalServerError, w),
			)
			return
		} else if schedules == nil {
			schedules = []*m.TestSchedule{}
		}

		switch r.Method {
		case http.MethodGet:
			w.Write(
				buildSuccess(schedules, w),
			)
		default:
			w.Write(buildFailure("Request not supported", http.StatusBadRequest, w))
		}
	}).Methods(http.MethodGet)

	// Delete a test schedule given a testscheduleid
	router.HandleFunc("/test-schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		scheduleid := vars["id"]

		// schedule exists?
		schedule, err := sto.GetTestSchedule(m.TestScheduleID(scheduleid))
		if err != nil {
			w.Write(
				buildFailure(err.Error(), http.StatusInternalServerError, w),
			)
			return
		} else if schedule == nil {
			w.Write(
				buildFailure(
					fmt.Sprintf("Cannot find TestSchedule<%s>", scheduleid),
					http.StatusNotFound,
					w,
				),
			)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			if err := server.sm.Remove(m.TestScheduleID(scheduleid)); err != nil {
				w.Write(
					buildFailure(err.Error(), http.StatusInternalServerError, w),
				)
				return
			}

			w.Write(
				buildSuccess(
					map[string]string{
						"scheduleid": scheduleid,
					},
					w,
				),
			)
		default:
			w.Write(buildFailure("Request not supported", http.StatusBadRequest, w))
		}
	}).Methods(http.MethodDelete)

	// Get grafana dashboard metadata
	router.HandleFunc("/dashboard-metadata", func(w http.ResponseWriter, r *http.Request) {
		if server.db == nil {
			w.Write(buildFailure("Grafana dashboards are not available", http.StatusNotFound, w))
			return
		}

		body, err := server.db.ToJSON()
		if err != nil {
			w.Write(buildFailure("Grafana dashboards are not available", http.StatusNotFound, w))
			return
		}
		w.Write(body)
	}).Methods(http.MethodGet)
}

// NewAPIServer create a new APIServer
func NewAPIServer(sched *sch.Scheduler, jf *mgr.JobFunnel, sm *mgr.ScheduleManager) *APIServer {
	db, _ := dash.NewDashboard()
	return &APIServer{sched, jf, sm, db}
}
