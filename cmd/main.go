package main

import (
	"fmt"
	"net/http"

	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/config"
	"github.com/t-bfame/diago/internal/scheduler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	config.Init()

	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption

	go func() {
		apiServer := server.NewAPIServer(s)
		apiServer.Start()
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%s", config.Diago.PrometheusPort), nil)
	}()

	server.InitGRPCServer("tcp", config.Diago.Host, config.Diago.GRPCPort, opts, s)
}
