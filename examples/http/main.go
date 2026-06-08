package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Chimera-State/go-otel-kit/middleware"
	"github.com/Chimera-State/go-otel-kit/setup"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	//tracing altyapısı
	ctx := context.Background()
	err := setup.Init(ctx,
		setup.WithServiceName("example-http-service"),
		setup.WithServiceVersion("1.0.0"),
		//deafult OTLP portumuz localhost:4317 di ve buna istek atacak
	)
	if err != nil {
		log.Fatalf("failed to init tracing: %v", err)
	}
	defer func() {
		if err := setup.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracing: %v", err)
		}
	}()

	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientReq, _ := http.NewRequestWithContext(r.Context(), "GET", "http://localhost:8080/world", nil)

		otel.GetTextMapPropagator().Inject(r.Context(), propagation.HeaderCarrier(clientReq.Header))

		//isteği gönder
		client := &http.Client{}
		resp, err := client.Do(clientReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf("Hello endpoint called another endpoint and received: %s", body)))
	})

	worldHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("World!"))
	})

	http.Handle("/hello", middleware.TraceMiddleware(helloHandler))
	http.Handle("/world", middleware.TraceMiddleware(worldHandler))

	fmt.Println("Server is running on http://localhost:8080...")
	fmt.Println("Test it by running: curl http://localhost:8080/hello")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
