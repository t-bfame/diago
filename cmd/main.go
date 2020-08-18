package main

import (
	"os"

	"github.com/t-bfame/diago/cmd/server"
	"github.com/t-bfame/diago/internal/scheduler"

	"google.golang.org/grpc"
)

func main() {
	s := scheduler.NewScheduler()
	var opts []grpc.ServerOption

	go func() {
		apiServer := server.NewApiServer()
		apiServer.Start()
	}()

	server.InitGRPCServer("tcp", os.Getenv("GRPC_HOST"), os.Getenv("GRPC_PORT"), opts, &s)
}
