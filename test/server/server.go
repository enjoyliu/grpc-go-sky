package main

import (
	"context"
	grpc_go_sky "grpc-go-sky"
	"log"
	"net"

	"github.com/SkyAPM/go2sky"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/SkyAPM/go2sky/reporter"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	oap         = "127.0.0.1:11800"
	port        = "127.0.0.1:50051"
	serviceName = "grpc-server"
	token = "this is a token"
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

	r, err := reporter.NewGRPCReporter(oap,reporter.WithAuthentication(token))
	if err != nil {
		panic(err)
	}
	defer r.Close()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serverName := serviceName
	tracer, err := go2sky.NewTracer(serverName, go2sky.WithReporter(r))
	if err != nil {
		panic(err)
	}
	grpcSvr := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_go_sky.UnaryServerInterceptor(tracer)),
	)

	pb.RegisterGreeterServer(grpcSvr, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcSvr.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Println("started!")
}
