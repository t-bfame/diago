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

func (s *workerServer) Register(stream pb.Worker_RegisterServer) error {

	// Expect Register message
	in, err := stream.Recv()

	if err != nil {
		log.Error("Encountered error: %v", err)
		return err
	}

	if in.Type != pb.Message_REGISTER {
		log.WithField("recvdType", fmt.Sprintf("%T", in.Type)).Error("Expected first message to be REGISTER, terminating connection")
		return errors.New("Expected first message to be REGISTER, terminating connection")
	}

	// TODO: figure out proper way of extracing message
	group := in.Group
	instance := scheduler.InstanceID(in.Instance)

	ch, err := s.sched.Register(group, instance)

	// Sending routine
	go func() {

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

	go func() {
		for {
			if err := stream.Send(nil); err != nil {
				break
			}
		}
	}()
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
