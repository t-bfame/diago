package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/t-bfame/diago/cmd/auth"
	dash "github.com/t-bfame/diago/pkg/dashboard"
	mgr "github.com/t-bfame/diago/pkg/manager"
	m "github.com/t-bfame/diago/pkg/model"
	sto "github.com/t-bfame/diago/pkg/storage"
	"golang.org/x/crypto/bcrypt"
)

// APIServer serves API calls over HTTP
type APIServer struct {
	jf mgr.JobFunnel
	sm mgr.ScheduleManager
	db *dash.Dashboard
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

func handleTestCreate(w http.ResponseWriter, r *http.Request) {
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
}

func handleTestReadForPrefix(w http.ResponseWriter, r *http.Request) {
	prefix := r.FormValue("prefix")
	tests, err := sto.GetAllTestsWithPrefix(prefix)
	if err != nil {
		w.Write(
			buildFailure(err.Error(), http.StatusInternalServerError, w),
		)
		return
	}
	w.Write(buildSuccess(tests, w))
}

func handleTestReadAll(w http.ResponseWriter, r *http.Request) {
	tests, err := sto.GetAllTests()
	if err != nil {
		w.Write(
			buildFailure(err.Error(), http.StatusInternalServerError, w),
		)
		return
	}
	w.Write(buildSuccess(tests, w))
}

func handleTestRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testid := vars["testid"]

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

	w.Write(buildSuccess(test, w))
}

func handleTestDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testid := vars["testid"]

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
}

func handleTestStartBuilder(
	server *APIServer,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["testid"]
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
	}
}

func handleTestStopBuilder(
	server *APIServer,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		testid := vars["testid"]
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
	}
}

func handleTestInstanceReadForTest(w http.ResponseWriter, r *http.Request) {
	testid := r.FormValue("testid")
	instances, err := sto.GetTestInstancesByTestIDWithLogs(m.TestID(testid))
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusInternalServerError, w))
		return
	}
	w.Write(buildSuccess(instances, w))
}

func handleTestInstanceReadAll(w http.ResponseWriter, r *http.Request) {
	tis, err := sto.GetAllTestInstancesWithLogs()
	if err != nil {
		w.Write(
			buildFailure(err.Error(), http.StatusInternalServerError, w),
		)
		return
	}
	w.Write(buildSuccess(tis, w))
}

func handleTestScheduleCreateBuilder(
	server *APIServer,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		err = m.Validate(reflect.TypeOf(m.TestSchedule{}), bodyContent)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		var schedule m.TestSchedule
		err = json.Unmarshal(bodyContent, &schedule)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		// make sure cron spec is valid
		err = server.sm.ValidateSpec(schedule.CronSpec)
		if err != nil {
			w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
			return
		}

		// make sure specified Test exists
		test, err := sto.GetTestByTestId(schedule.TestID)
		if err != nil {
			w.Write(
				buildFailure(err.Error(), http.StatusInternalServerError, w),
			)
			return
		} else if test == nil {
			w.Write(buildFailure(
				fmt.Sprintf("Cannot find Test<%s>", schedule.TestID),
				http.StatusBadRequest,
				w,
			))
			return
		}

		schedule.ID = m.TestScheduleID(schedule.Name)
		if err := server.sm.Add(&schedule, true); err != nil {
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
	}
}

func handleTestScheduleReadForTest(w http.ResponseWriter, r *http.Request) {
	testid := r.FormValue("testid")

	schedules, err := sto.GetTestSchedulesByTestID(m.TestID(testid))
	if err != nil {
		w.Write(
			buildFailure(err.Error(), http.StatusInternalServerError, w),
		)
		return
	}

	w.Write(
		buildSuccess(schedules, w),
	)
}

func handleTestScheduleReadAll(w http.ResponseWriter, r *http.Request) {
	schedules, err := sto.GetAllTestSchedules()
	if err != nil {
		w.Write(
			buildFailure(err.Error(), http.StatusInternalServerError, w),
		)
		return
	}
	w.Write(buildSuccess(schedules, w))
}

func handleTestScheduleDeleteBuilder(
	server *APIServer,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		scheduleid := vars["scheduleid"]

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
	}
}

func handleLoginUser(w http.ResponseWriter, r *http.Request) {
	bodyContent, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
		return
	}

	err = m.Validate(reflect.TypeOf(m.User{}), bodyContent)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
		return
	}

	var user m.User
	err = json.Unmarshal(bodyContent, &user)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
		return
	}

	foundUser, err := sto.GetUserByUserId(m.UserID(user.Username))

	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusForbidden, w))
		return
	}

	if !CheckPasswordHash(user.Password, foundUser.Password) {
		w.Write(buildFailure(err.Error(), http.StatusForbidden, w))
		return
	}

	token, err := auth.GenerateToken(user.Username)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusInternalServerError, w))
		return
	}

	w.Write(
		buildSuccess(
			map[string]string{
				"token": token,
			},
			w,
		),
	)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	bodyContent, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
		return
	}

	err = m.Validate(reflect.TypeOf(m.User{}), bodyContent)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
		return
	}

	var user m.User
	err = json.Unmarshal(bodyContent, &user)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusBadRequest, w))
		return
	}

	err = sto.AddUser(&user)

	token, err := auth.GenerateToken(user.Username)
	if err != nil {
		w.Write(buildFailure(err.Error(), http.StatusInternalServerError, w))
		return
	}

	w.Write(
		buildSuccess(
			map[string]string{
				"token": token,
			},
			w,
		),
	)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Start starts the APIServer
func (server *APIServer) Start(router *mux.Router) {
	router.Use(preResponse)
	router.Use(auth.Middleware())
	// consider using:
	// router.Use(auth.Middleware(), cors.Default().Handler)
	// cors from "github.com/rs/cors"

	// users
	router.HandleFunc("/user", handleCreateUser).Methods(http.MethodPost)
	router.HandleFunc("/login", handleLoginUser).Methods(http.MethodPost)

	// tests
	router.HandleFunc("/tests", handleTestCreate).Methods(http.MethodPost)
	router.HandleFunc("/tests", handleTestReadForPrefix).Methods(http.MethodGet).
		Queries("prefix", "{prefix}")
	router.HandleFunc("/tests", handleTestReadAll).Methods(http.MethodGet)
	router.HandleFunc("/tests/{testid}", handleTestRead).Methods(http.MethodGet)
	router.HandleFunc("/tests/{testid}", handleTestDelete).Methods(http.MethodDelete)
	router.HandleFunc("/tests/{testid}/start", handleTestStartBuilder(server)).
		Methods(http.MethodPost)
	router.HandleFunc("/tests/{testid}/stop", handleTestStopBuilder(server)).
		Methods(http.MethodPost)

	// test-instances
	router.HandleFunc("/test-instances", handleTestInstanceReadForTest).
		Methods(http.MethodGet).Queries("testid", "{testid}")
	router.HandleFunc("/test-instances", handleTestInstanceReadAll).Methods(http.MethodGet)

	// test-schedules
	router.HandleFunc("/test-schedules", handleTestScheduleCreateBuilder(server)).
		Methods(http.MethodPost)
	router.HandleFunc("/test-schedules", handleTestScheduleReadForTest).
		Methods(http.MethodGet).Queries("testid", "{testid}")
	router.HandleFunc("/test-schedules", handleTestScheduleReadAll).
		Methods(http.MethodGet)
	router.HandleFunc("/test-schedules/{scheduleid}", handleTestScheduleDeleteBuilder(server)).
		Methods(http.MethodDelete)

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
func NewAPIServer(jf mgr.JobFunnel, sm mgr.ScheduleManager) *APIServer {
	db, _ := dash.NewDashboard()
	return &APIServer{jf, sm, db}
}
