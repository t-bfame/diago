/*
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	pb "github.com/t-bfame/diago/proto-gen/lol"
	"google.golang.org/grpc"
)

type routeGuideServer struct {
	pb.UnimplementedRouteGuideServer
}

func randomPoint(r *rand.Rand) *pb.Point {
	lat := (r.Int31n(180) - 90) * 1e7
	long := (r.Int31n(360) - 180) * 1e7
	return &pb.Point{Latitude: lat, Longitude: long}
}

func (s *routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		for {
			if err := stream.Send(randomPoint(r)); err != nil {
				break
			}
		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Println(in)
	}
}

func newServer() *routeGuideServer {
	s := &routeGuideServer{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
*/