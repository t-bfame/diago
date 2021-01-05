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
		log.WithError(err).Error("Encountered error during registration")
		return err
	}

	if reg := msg.GetRegister(); reg == nil {
		log.WithField("recvdType", fmt.Sprintf("%T", msg.Payload)).Error("Expected first message to be Register, terminating connection")
		return errors.New("Expected first message to be Register, terminating connection")
	}

	reg := msg.GetRegister()
	group := reg.GetGroup()
	freq := reg.GetFrequency()

	instance := scheduler.InstanceID(reg.GetInstance())

	log.WithField("group", group).WithField("instance", instance).WithField("frequency", freq).Info("Received registration for pod")

	leaderMsgs, workerMsgs, err := s.sched.Register(group, instance, freq)

	if err != nil {
		log.WithError(err).Error("Encountered error during registeration")
		return err
	}

	// Sending routine
	go func() {
		for event := range workerMsgs {
			if err := stream.Send(event.ToProto()); err != nil {
				log.WithError(err).Error("Error sending message to worker")
			}
		}
	}()

	// Receiving routine
	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.WithError(err).WithField("group", group).WithField("instance", instance).Error("Encountered receiver stream error")
			break
		}

		inc, err := scheduler.ProtoToIncoming(msg)
		if err != nil {
			log.WithField("recvdType", fmt.Sprintf("%T", msg.Payload)).Error("Enountered messsage with unexpected type, discarding message")
		}

		leaderMsgs <- inc
	}

	log.WithField("group", group).WithField("instance", instance).Info("Closing pod")
	close(leaderMsgs)

	return nil
}

func newServer(s *scheduler.Scheduler) *workerServer {
	return &workerServer{sched: s}
}

// InitGRPCServer Initializes the gRPC server for diago
func InitGRPCServer(protocol string, host string, port string, opts []grpc.ServerOption, s *scheduler.Scheduler) {

	lis, err := net.Listen(protocol, fmt.Sprintf("%s:%s", host, port))

	if err != nil {
		log.Fatalf("gRPC server failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterWorkerServer(grpcServer, newServer(s))
	defer grpcServer.Serve(lis)

	log.WithField("host", host).WithField("port", port).Info("gRPC server listening")
}
