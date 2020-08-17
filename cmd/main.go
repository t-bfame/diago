package main

import (
	// "fmt"
	"os"
	// "time"

	"github.com/t-bfame/diago/cmd/server"
	// "github.com/t-bfame/diago/internal/manager"
	"github.com/t-bfame/diago/internal/scheduler"
	"github.com/t-bfame/diago/api/apiserver"

	"google.golang.org/grpc"
)

func main() {
	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption

	// go func() {
	// 	time.Sleep(5 * time.Second)

	// 	fmt.Println("Submitting")

	// 	ti := manager.Job{
	// 		ID:         "1",
	// 		Name:       "alpha",
	// 		Group:      "diago-worker",
	// 		Priority:   0,
	// 		Frequency:  10,
	// 		Duration:   10,
	// 		HTTPMethod: "GET",
	// 		HTTPUrl:    "https://www.google.com",
	// 	}

	// 	ch, _ := s.Submit(ti)

	// 	for msg := range ch {
	// 		fmt.Println(msg)
	// 		fmt.Println(fmt.Sprintf("%T", msg))
	// 	}

	// 	fmt.Println("DONE")

	// }()

	go func() {
		apiServer := apiserver.NewApiServer()
		apiServer.Start()
	}()

	server.InitGRPCServer("tcp", os.Getenv("GRPC_HOST"), os.Getenv("GRPC_PORT"), opts, &s)
}
