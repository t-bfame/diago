package server

import (
	"errors"
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/t-bfame/diago/pkg/scheduler"
	pb "github.com/t-bfame/diago/proto-gen/aggregator"
	"google.golang.org/grpc"
)

type aggregatorServer struct {
	pb.UnimplementedAggregatorServer

	sched *scheduler.Scheduler
}

// client-side streaming to receive aggregated metrics
func (s *aggregatorServer) Coordinate(stream pb.Aggregator_CoordinateServer) error {

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

	log.Info("Received registration for aggregator pod")

	incomingMsgs, err := s.sched.RegisterAgg()

	if err != nil {
		log.WithError(err).Error("Encountered error during registration")
		return err
	}

	// Receiving routine
	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Encountered aggregator receiver stream error")
			break
		}

		inc, err := scheduler.ProtoToIncomingAgg(msg)
		if err != nil {
			log.Error("Encountered aggregator messsage with unexpected type, discarding message")
			break
		}

		incomingMsgs <- inc
	}

	log.Info("Closing aggregator pod")
	close(incomingMsgs)

	return nil
}

func newAggServer(s *scheduler.Scheduler) *aggregatorServer {
	return &aggregatorServer{sched: s}
}

// InitGRPCServer Initializes the gRPC server for diago
func InitAggregatorGRPCServer(protocol string, host string, port uint64, opts []grpc.ServerOption, s *scheduler.Scheduler) {

	lis, err := net.Listen(protocol, fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		log.Fatalf("gRPC server failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterAggregatorServer(grpcServer, newAggServer(s))
	defer grpcServer.Serve(lis)

	log.WithField("host", host).WithField("port", port).Info("gRPC server listening")
}
