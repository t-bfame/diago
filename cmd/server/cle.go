/*
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
	"time"

	pb "github.com/t-bfame/diago/proto-gen/lol"
	"google.golang.org/grpc"
)

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
func runRouteChat(client pb.RouteGuideClient) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.RouteChat(ctx)
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()

			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}

			log.Printf("Got message at point(%d, %d)", in.Latitude, in.Longitude)
		}
	}()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		if err := stream.Send(randomPoint(r)); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}

		time.Sleep(time.Second)
	}

	stream.CloseSend()
	<-waitc
}

func randomPoint(r *rand.Rand) *pb.Point {
	lat := (r.Int31n(180) - 90) * 1e7
	long := (r.Int31n(360) - 180) * 1e7
	return &pb.Point{Latitude: lat, Longitude: long}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial("localhost:8000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewRouteGuideClient(conn)

	// RouteChat
	runRouteChat(client)
}
*/