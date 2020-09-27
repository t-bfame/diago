package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	mgr "github.com/t-bfame/diago/internal/manager"
	m "github.com/t-bfame/diago/internal/model"
	sch "github.com/t-bfame/diago/internal/scheduler"
)

type ApiServer struct {
	scheduler *sch.Scheduler
	funnel    *mgr.JobFunnel
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
		log.Info(fmt.Sprintf("Failed to build response for %v", respMap))
		return make([]byte, 0)
	}
	return json
}

// Start starts the ApiServer
func (server *ApiServer) Start() {
	router := mux.NewRouter()
	router.Use(preResponse)

	// Test C
	router.HandleFunc("/tests", func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		test, err := m.SaveTestFromBody(bodyContent, true)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		w.Write(buildSuccess(test, w))
	}).Methods(http.MethodPost)

	// Test RUD
	router.HandleFunc("/tests/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]
		test, found := m.TestByID(m.TestID(testid))
		if !found {
			w.Write(buildFailure("Test not found", http.StatusNotFound, w))
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Write(buildSuccess(test, w))
		case http.MethodPut:
			bodyContent, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
				return
			}

			test, err := m.SaveTestFromBody(bodyContent)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
				return
			}

			w.Write(buildSuccess(test, w))
		case http.MethodDelete:
			ok, err := test.Delete()
			if !ok {
				w.Write(
					buildFailure(err.Error(), http.StatusBadRequest, w),
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
		instances, found := m.TestInstancesByTestID(m.TestID(testid))
		if !found {
			w.Write(
				buildFailure(
					fmt.Sprintf("No instances found for Test<%s>", testid),
					http.StatusNotFound,
					w,
				),
			)
			return
		}

		w.Write(buildSuccess(instances, w))
	}).Methods(http.MethodGet)

	// Start adhoc
	router.HandleFunc("/start-test/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]
		ok, err := server.funnel.BeginTest(m.TestID(testid))
		if !ok {
			w.Write(buildFailure(err.Error(), http.StatusNotFound, w))
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
		ok, err := server.funnel.StopTest(m.TestID(testid))
		if !ok {
			w.Write(buildFailure(err.Error(), http.StatusNotFound, w))
			return
		}

		w.Write(
			buildSuccess(
				fmt.Sprintf("Successfully stopped Test<%s>", testid),
				w,
			),
		)
	})

	port := os.Getenv("API_PORT")
	defer http.ListenAndServe(":"+port, router)
	log.WithField("port", port).Info("Api server listening")
}

func NewApiServer(sched *sch.Scheduler, funnel *mgr.JobFunnel) *ApiServer {
	return &ApiServer{
		sched,
		funnel,
	}
}
