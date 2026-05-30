package grpc

import (
	"context"
	"fmt"
	"log"

	"github.com/Chimera-State/go-otel-kit/interceptor"
	"github.com/Chimera-State/go-otel-kit/setup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// tracing başlar
	ctx := context.Background()
	err := setup.Init(ctx, setup.WithServiceName("example-grpc-client"))
	if err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}
	defer setup.Shutdown(ctx)

	opts := append(interceptor.ClientOptions(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// burada normalde proto'dan generate olan client kodu
	// örnek amaçlı sadece trace contextin başarılı bir şekilde incejt edildiğini
	// göstermek için client bağlantısnı kuruyoz.
	fmt.Println("gRPC Client successfully connected and ready to trace context analysis!")
}
