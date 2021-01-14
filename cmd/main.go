package main

import (
	"fmt"
	"net/http"

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
	storage.InitDatabase(storage.DatabaseName)

	if config.Diago.Debug {
		log.SetLevel(log.DebugLevel)
	}

	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption

	go func() {
		apiServer := server.NewAPIServer(s)
		apiServer.Start()
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", config.Diago.PrometheusPort), nil)
	}()

	server.InitGRPCServer("tcp", config.Diago.Host, config.Diago.GRPCPort, opts, s)
}
