package apiserver

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"io/ioutil"
	"net/http"
	"encoding/json"
	mgr "github.com/t-bfame/diago/internal/manager"
	sch "github.com/t-bfame/diago/internal/scheduler"
	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

type ApiServer struct {
	scheduler *sch.Scheduler
	dummyTests map[string]mgr.Test
	dummyInstances map[string][]*mgr.TestInstance
	ongoingTests map[string]bool
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
		// ???
		return make([]byte, 0)
	}
	return json
}

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
		var test mgr.Test;
		err = json.Unmarshal(bodyContent, &test)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		// For now, make id by concatenating test name to timestamp
		uid := test.Name + "-" + strconv.FormatInt(time.Now().Unix(), 10)
		test.ID = mgr.TestID(uid)
		server.dummyTests[uid] = test

		w.Write(
			buildSuccess(
				map[string]string{
					"testid": uid,
				},
				w,
			),
		)

		log.WithField("TestID", uid).Info("Test created")

	}).Methods(http.MethodPost)

	// Test RUD
	router.HandleFunc("/tests/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		test, found := server.dummyTests[id]
		if !found {
			w.Write(buildFailure("Test not found", http.StatusNotFound, w))
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Write(buildSuccess(test, w))

			log.WithField("TestID", id).Info("Test retrieved")
		case http.MethodPut:
			bodyContent, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
				return
			}
			var updatedTest mgr.Test;
			err = json.Unmarshal(bodyContent, &updatedTest)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
				return
			}
			server.dummyTests[id] = updatedTest
			w.Write(buildSuccess(updatedTest, w))

			log.WithField("TestID", id).Info("Test updated")
		case http.MethodDelete:
			delete(server.dummyTests, id)
			w.Write(
				buildSuccess(
					map[string]string{
						"testid": id,
					},
					w,
				),
			)

			log.WithField("TestID", id).Info("Test deleted")
		default:
			w.Write(buildFailure("Request not supported", http.StatusBadRequest, w))
		}
	}).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)

	// Get TestInstances by TestID
	router.HandleFunc("/test-instances/{testid}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["testid"]
		instances, found := server.dummyInstances[testid]
		if !found {
			w.Write(buildFailure("Test not found", http.StatusNotFound, w))
			return
		}
		w.Write(buildSuccess(instances, w))

		log.WithField("TestID", testid).Info("Test instances retrieved")
	}).Methods(http.MethodGet)

	// Start adhoc
	router.HandleFunc("/start-test/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]
		test, found := server.dummyTests[testid]
		if !found {
			w.Write(buildFailure("Test not found", http.StatusNotFound, w))
			return
		}

		// Check if there's an instance of this test in progress already
		_, exists := server.ongoingTests[testid]
		if exists {
			w.Write(buildFailure("Test already ongoing", http.StatusConflict, w))
			return
		}

		// create TestInstance
		// for now generate uid using test name + timestamp
		now := time.Now().Unix()
		uid := test.Name + "-" + strconv.FormatInt(now, 10)
		testInstance := &mgr.TestInstance{
			ID: mgr.TestInstanceID(uid),
			TestID: mgr.TestID(testid),
			Type: "adhoc",
			Status: "submitted",
			CreatedAt: now,
		}
		server.dummyInstances[testid] = append(
			server.dummyInstances[testid],
			testInstance,
		)

		for _, v := range test.Jobs {
			// submit jobs to scheduler and listen
			// on the channel until termination
			ch, err := server.scheduler.Submit(v)
			if err != nil {
				testInstance.Status = "failed"
				w.Write(buildFailure(err.Error(), http.StatusConflict, w))
				// TODO: probably also stop every other submitted job here
				return
			}

			// monitor the channel
			// TODO: stream to metrics aggregator
			go func() {
				for msg := range ch {
					switch {
					default:
						fmt.Println(msg)
					}
				}
				// TODO: write aggregated metrics to TestInstance
				testInstance.Status = "done"
				// Remove from ongoingTests when channel closes
				delete(server.ongoingTests, testid)
				log.
					WithField("TestID", testid).
					WithField("TestInstanceID", uid).
					WithField("JobID", v.ID).
					Info("Finished/Stopped Job")
			}()
		}

		server.ongoingTests[testid] = true

		w.Write(
			buildSuccess(
				map[string]string {
					"testid": testid,
				},
				w,
			),
		)

		log.
			WithField("TestID", testid).
			WithField("TestInstanceID", uid).
			Info("Test submitted")
	})

	// stop TestInstance by TestID
	router.HandleFunc("/stop-test/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["id"]
		_, found := server.ongoingTests[testid]
		if !found {
			w.Write(buildFailure("No such ongoing Test exists.", http.StatusNotFound, w))
			return
		}

		// stop all jobs
		for _, v := range server.dummyTests[testid].Jobs {
			err := server.scheduler.Stop(v)
			if err != nil {
				w.Write(buildFailure(err.Error(), http.StatusInternalServerError, w))
				return
			}
		}

		delete(server.ongoingTests, testid)

		w.Write(
			buildSuccess(
				map[string]string {
					"testid": testid,
				},
				w,
			),
		)

		log.WithField("TestID", testid).Info("Test stopped")
	})
	
	port := os.Getenv("API_PORT")
	defer http.ListenAndServe(":" + port, router)
	log.WithField("port", port).Info("Api server listening")
}

func NewApiServer() *ApiServer {
	scheduler := sch.NewScheduler()
	return &ApiServer{
		&scheduler,
		make(map[string]mgr.Test),
		make(map[string][]*mgr.TestInstance),
		make(map[string]bool),
	}
}