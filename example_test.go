package grpc_go_sky

import (
	"context"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	oap         = "mockoap:19876"
	port        = ":50051"
	serviceName = "grpc-server"
)

func Test(t *testing.T) {
	//Use gRPC reporter for production
	r, err := reporter.NewLogReporter()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	// new server tracer
	serverName := serviceName
	tracer, err := go2sky.NewTracer(serverName, go2sky.WithReporter(r))
	if err != nil {
		panic(err)
	}
	// run server
	go func() {
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcSvr := grpc.NewServer(grpc_middleware.WithUnaryServerChain(
			UnaryServerInterceptor(tracer),
		))
		if err = grpcSvr.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// run client
	time.Sleep(5 * time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// new client tracer
		clientName := "grpc-client"
		cliTracer, err := go2sky.NewTracer(clientName, go2sky.WithReporter(r))
		if err != nil {
			panic(err)
		}
		opts := grpc.WithChainUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			UnaryClientInterceptor(cliTracer)))
		grpcCli, err := grpc.DialContext(
			context.Background(),
			"",
			opts,
		)
		if err != nil {
			panic(err)
		}
		// new client
		grpcClient := pb.NewGreeterClient(grpcCli)
		// send hello
		resp, err := grpcClient.SayHello(context.Background(), &pb.HelloRequest{Name: "grpc-hello"})
		if err != nil {
			panic(err)
		}
		t.Logf("[grpc] Say hello: %s\n", resp)
	}()
	wg.Wait()

}
