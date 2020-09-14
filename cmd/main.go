package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/internal/scheduler"
	// "github.com/t-bfame/diago/internal/model"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption

	// model.InitModel()
	// body1 := []byte(`{
	// 	"Name":"Test1",
	// 	"Jobs":[
	// 		{
	// 			"Name":"alpha",
	// 			"Group":"diago-worker",
	// 			"Priority":1,
	// 			"Frequency":5,
	// 			"Duration":30,
	// 			"HTTPMethod":"GET",
	// 			"HTTPUrl":"https://www.google.com"
	// 		}
	// 	]
	// }`)
	// content, err := model.TestFromBody(body1)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// test, err := model.CreateTest(content)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(test)
	// fmt.Println(model.TestById(test["ID"].(model.TestID)))

	// body2 := []byte(`{
	// 	"Name":"Test2",
	// 	"Jobs":[
	// 		{
	// 			"Name":"alpha",
	// 			"Group":"diago-worker",
	// 			"Priority":0,
	// 			"Frequency":4242,
	// 			"Duration":24242,
	// 			"HTTPMethod":"POST",
	// 			"HTTPUrl":"https://www.github.com"
	// 		}
	// 	]
	// }`)
	// content1, err := model.TestFromBody(body2)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// test1, err := model.CreateTest(content1)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(test1)
	// model.DumpStorage()

	go func() {
		apiServer := server.NewApiServer(&s)
		apiServer.Start()
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PROMETHEUS_PORT")), nil)
	}()

	server.InitGRPCServer("tcp", os.Getenv("GRPC_HOST"), os.Getenv("GRPC_PORT"), opts, &s)
}
