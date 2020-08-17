package server

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/t-bfame/diago/internal/scheduler"

	log "github.com/sirupsen/logrus"

	pb "github.com/t-bfame/diago/proto-gen/worker"
	"google.golang.org/grpc"
)

type workerServer struct {
	pb.UnimplementedWorkerServer

	sched *scheduler.Scheduler
}

func (s *workerServer) Coordinate(stream pb.Worker_CoordinateServer) error {

	// Expect Register message
	msg, err := stream.Recv()

	if err != nil {
		log.Error("Encountered error: %v", err)
		return err
	}

	if msg.Payload.(type) != *pb.Message_Register {
		log.WithField("recvdType", fmt.Sprintf("%T", msg.Payload.(type))).Error("Expected first message to be REGISTER, terminating connection")
		return errors.New("Expected first message to be Register, terminating connection")
	}

	// TODO: figure out proper way of extracing message
	group := msg.Payload.Group
	instance := scheduler.InstanceID(msg.Payload.Group.Instance)

	ch, err := s.sched.Register(group, instance)

	// Sending routine
	go func() {
		for {
			if err := stream.Send(nil); err != nil {
				break
			}
		}

	}()

	// Receiving routine
	for {
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Println(in)
	}
}

func newServer(s *scheduler.Scheduler) *workerServer {
	return &workerServer{sched: s}
}

// InitGRPCServer Initializes the gRPC server for diago
func InitGRPCServer(protocol string, host string, port uint16, opts []grpc.ServerOption, s *scheduler.Scheduler) {

	lis, err := net.Listen(protocol, fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		log.Fatalf("gRPC server failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterWorkerServer(grpcServer, newServer(s))
	defer grpcServer.Serve(lis)

	log.WithField("host", host).WithField("port", port).Info("gRPC server listening")
}
