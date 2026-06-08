package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Chimera-State/go-otel-kit/interceptor"
	"github.com/Chimera-State/go-otel-kit/setup"
	"google.golang.org/grpc"
)

func main() {
	// tracing başlar
	ctx := context.Background()
	err := setup.Init(ctx, setup.WithServiceName("example-grpc-server"))
	if err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}
	defer func() {
		if err := setup.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracing: %v", err)
		}
	}()

	// gRPC serveri oluştur
	server := grpc.NewServer(interceptor.ServerOptions()...)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fmt.Println("gRPC Server is running on :50051...")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
