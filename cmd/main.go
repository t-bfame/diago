package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/config"
	"github.com/t-bfame/diago/pkg/chaosmgr"
	"github.com/t-bfame/diago/pkg/manager"
	"github.com/t-bfame/diago/pkg/scheduler"
	"github.com/t-bfame/diago/pkg/storage"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	config.Init()

	if err := storage.InitDatabase(config.Diago.StoragePath); err != nil {
		panic("Failed to init database.")
	}

	mongoPath := fmt.Sprintf("mongodb://%s:%d", config.Diago.MongoDBHost, config.Diago.MongoDBPort)
	if err := storage.ConnectToMongoDB(context.Background(), mongoPath); err != nil {
		panic(fmt.Sprintf("Failed to init mongo at %s.", mongoPath))
	}

	if config.Diago.Debug {
		log.SetLevel(log.DebugLevel)
	}

	s := scheduler.NewScheduler()
	cm := chaosmgr.NewChaosManager()
	var opts []grpc.ServerOption

	router := mux.NewRouter()

	go func() {
		jf := manager.NewJobFunnel(s, cm)
		sm := manager.NewScheduleManager(jf)

		// Set prefix for api paths
		apiRouter := router.PathPrefix("/api").Subrouter()
		apiServer := server.NewAPIServer(jf, sm)
		apiServer.Start(apiRouter)

		server.NewUIBox(router)

		defer http.ListenAndServe(fmt.Sprintf(":%d", config.Diago.APIPort), router)
		log.WithField("port", config.Diago.APIPort).Info("Api server listening")
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", config.Diago.PrometheusPort), nil)
	}()

	server.InitGRPCServer("tcp", config.Diago.Host, config.Diago.GRPCPort, opts, s)
}
