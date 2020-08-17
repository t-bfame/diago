package main

import (
	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/internal/scheduler"

	"google.golang.org/grpc"
)

func main() {
	// ti := manager.Job{
	// 	ID:       "1",
	// 	Name:     "alpha",
	// 	Group:    "diago-worker",
	// 	Priority: 0,
	// }

	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption
	server.InitGRPCServer("tcp", "localhost", 5000, opts, &s)
}
