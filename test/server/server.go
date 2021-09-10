package main

import (
	"context"
	"log"
	"net"

	"github.com/SkyAPM/go2sky/reporter"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	oap         = "mockoap:19876"
	port        = ":50051"
	serviceName = "g-server"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {

	r, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcSvr := grpc.NewServer()

	pb.RegisterGreeterServer(grpcSvr, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcSvr.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
