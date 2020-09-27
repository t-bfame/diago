package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/internal/manager"
	"github.com/t-bfame/diago/internal/model"
	"github.com/t-bfame/diago/internal/scheduler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption

	model.InitModel()
	jf := manager.NewJobFunnel(&s)

	go func() {
		apiServer := server.NewApiServer(&s, jf)
		apiServer.Start()
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PROMETHEUS_PORT")), nil)
	}()

	server.InitGRPCServer("tcp", os.Getenv("GRPC_HOST"), os.Getenv("GRPC_PORT"), opts, &s)
}
