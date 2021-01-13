package main

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/t-bfame/diago/api/v1alpha1"
	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/config"
	"github.com/t-bfame/diago/internal/manager"
	"github.com/t-bfame/diago/internal/model"
	"github.com/t-bfame/diago/internal/scheduler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	config.Init()

	if config.Diago.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// create the in-cluster config
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}

	// create CRD client
	crdClient, err := v1alpha1.NewClient(restConfig)
	if err != nil {
		panic(err.Error())
	}

	s := scheduler.NewScheduler(clientset, crdClient)

	tests := make(map[string]model.Test)
	instances := make(map[string][]*model.TestInstance)
	jf := manager.NewJobFunnel(s)
	go jf.PrepareScheduledTests(crdClient, &tests, &instances)

	var opts []grpc.ServerOption

	go func() {
		apiServer := server.NewAPIServer(s, jf, &tests, &instances)
		apiServer.Start()
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", config.Diago.PrometheusPort), nil)
	}()

	server.InitGRPCServer("tcp", config.Diago.Host, config.Diago.GRPCPort, opts, s)
}
