package main

import (
	"fmt"
	"io"
	"log"
	"net"

	worker "github.com/t-bfame/diago/proto-gen/worker"

	"google.golang.org/grpc"
)

type workerServer struct {
	worker.UnimplementedWorkerServer
}

func (s *workerServer) Coordinate(stream worker.Worker_CoordinateServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Client has ended connection")
			return nil
		}
		if err != nil {
			return err
		}
		switch x := msg.Payload.(type) {
		case *worker.Message_Register:
			fmt.Println("Received message of type: Register")

			httpRequest := &worker.HTTPRequest{
				Method: "GET",
				Url:    "http://localhost:3000",
			}

			msgStart := &worker.Message{
				Payload: &worker.Message_Start{
					Start: &worker.Start{
						Jobid:     0,
						Frequency: 1,
						Duration:  3,
						Request:   httpRequest,
					},
				},
			}
			if err := stream.Send(msgStart); err != nil {
				fmt.Println("Error sending start message!")
				return err
			}
		case *worker.Message_Metrics:
			fmt.Println("Received message of type: Metrics")
			metrics := x.Metrics
			fmt.Printf("Latency is: %v\n", metrics.GetLatency())
		case *worker.Message_Finish:
			fmt.Println("A worker has finished processing")
		case nil:
			// The field is not set.
		default:
			return fmt.Errorf("Message.Payload has unexpected type %T", x)
		}
	}
}

func newServer() *workerServer {
	s := &workerServer{}
	return s
}

func main() {
	fmt.Println("Starting server!")

	// TODO: change the port
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 5000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	worker.RegisterWorkerServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
