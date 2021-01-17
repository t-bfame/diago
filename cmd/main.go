package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/config"
	"github.com/t-bfame/diago/internal/scheduler"
	"github.com/t-bfame/diago/internal/storage"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	config.Init()
	storage.InitDatabase(config.Diago.StoragePath)

	if config.Diago.Debug {
		log.SetLevel(log.DebugLevel)
	}

	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption

	router := mux.NewRouter()

	go func() {
		// Set prefix for api paths
		apiRouter := router.PathPrefix("/api").Subrouter()
		apiServer := server.NewAPIServer(s)
		apiServer.Start(apiRouter)
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", config.Diago.PrometheusPort), nil)
	}()

	server.InitGRPCServer("tcp", config.Diago.Host, config.Diago.GRPCPort, opts, s)
}
