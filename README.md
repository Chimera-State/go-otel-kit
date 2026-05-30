[English](README.md) | [Türkçe](README.tr.md)

# go-otel-kit

Go uygulamaları için OpenTelemetry tabanlı distributed tracing araç kiti.  
HTTP, gRPC, Kafka, Redis ve PostgreSQL üzerinden geçen isteklerin trace'lerini Jaeger ve Grafana üzerinde görselleştirir.

---

## 🚀 Observability Stack Setup

Projeye yeni katılan biri için tam stack'i **tek komutla** ayağa kaldırma rehberi.  
Aşağıdaki adımları izleyerek 10 dakika içinde çalışan bir observability ortamına sahip olabilirsin.

### Gereksinimler

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) kurulu ve çalışıyor olmalı
- [Go 1.21+](https://go.dev/dl/) kurulu olmalı (örnek servisleri çalıştırmak için)
- Git ile repo klonlanmış olmalı

---

### 1. Stack'i Başlat

```bash
docker compose -f docker-compose.full.yml up -d
```

Bu komut aşağıdaki servisleri tek seferde ayağa kaldırır:

| Servis         | Açıklama                                      |
|----------------|-----------------------------------------------|
| Jaeger         | Distributed trace görselleştirme              |
| OTel Collector | Trace toplama ve yönlendirme                  |
| Grafana        | Dashboard ve metrik görselleştirme            |
| Kafka          | Asenkron mesajlaşma altyapısı (KRaft modu)    |
| PostgreSQL     | İlişkisel veritabanı                          |
| Redis          | Cache                                         |

---

### 2. Servislerin Durumunu Kontrol Et

```bash
docker compose -f docker-compose.full.yml ps
```

Tüm servisler `Up` durumunda görünmelidir.

---

### 3. Port Listesi

| Servis         | Port  | Açıklama                          |
|----------------|-------|-----------------------------------|
| Jaeger UI      | 16686 | Trace görselleştirme arayüzü      |
| OTel Collector | 4317  | OTLP gRPC uygulama bağlantı noktası |
| OTel Collector | 4318  | OTLP HTTP uygulama bağlantı noktası |
| OTel Collector | 8888  | Collector kendi metrikleri        |
| Grafana        | 3000  | Dashboard arayüzü                 |
| Kafka          | 9092  | Kafka broker                      |
| PostgreSQL     | 5432  | Veritabanı                        |
| Redis          | 6379  | Cache                             |

---

### 4. Arayüzlere Erişim

**Jaeger** → http://localhost:16686  
Kullanıcı adı veya şifre gerekmez.

**Grafana** → http://localhost:3000  
Kullanıcı: `admin` / Şifre: `admin`

---

### 5. İlk Trace'i Görmek

Stack çalışırken örnek HTTP servisini başlat:

```bash
cd examples/http
go run main.go
```

Yeni bir terminalde test isteği at:

```bash
curl http://localhost:8080/hello
```

**Jaeger'da trace'i incele:**

1. http://localhost:16686 adresini aç
2. Sol menüden **Service:** `example-http-service` seç
3. **Find Traces** butonuna tıkla
4. Listelenen trace'e tıkla → HTTP span zincirini göreceksin

---

### 6. Stack'i Durdurma

```bash
docker compose -f docker-compose.full.yml down
```

---

## 🔧 Backend Integration Guide

This library provides ready-to-use tools that make it easy to set up observability (tracing) in your microservice architectures (HTTP, gRPC, Kafka).

### Installation

```bash
go get github.com/Chimera-State/go-otel-kit
```

### Provider Initialization

```go
package main

import (
	"context"
	"log"

	"github.com/Chimera-State/go-otel-kit/setup"
)

func main() {
	ctx := context.Background()

	err := setup.Init(ctx,
		setup.WithServiceName("my-backend-service"),
		setup.WithServiceVersion("1.0.0"),
		setup.WithExporterEndpoint("localhost:4317"),
		setup.WithSamplingRate(1.0),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Tracer: %v", err)
	}
	defer setup.Shutdown(ctx)
}
```

### HTTP Middleware

```go
tracedHandler := middleware.TraceMiddleware(mux)
http.ListenAndServe(":8080", tracedHandler)
```

### gRPC Interceptor

```go
grpcServer := grpc.NewServer(
	grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor()),
)
```

### Kafka Inject & Extract

```go
// PRODUCER
func PublishEvent(ctx context.Context, client *kgo.Client) {
	record := &kgo.Record{Topic: "orders-topic", Value: []byte("new-order")}
	kafka.InjectToRecord(ctx, record)
	client.Produce(ctx, record, nil)
}

// CONSUMER
func ConsumeEvent(client *kgo.Client) {
	ctx := context.Background()
	fetches := client.PollFetches(ctx)
	iter := fetches.RecordIter()
	for !iter.Done() {
		record := iter.Next()
		extractedCtx := kafka.ExtractFromRecord(ctx, record)
		tracer := otel.Tracer("kafka-consumer")
		_, span := tracer.Start(extractedCtx, "process-order-event")
		// business logic...
		span.End()
	}
}
```

### PostgreSQL (GORM) & Redis

```go
// Redis
redisotel.InstrumentTracing(rdb)

// GORM
db.Use(tracing.NewPlugin())
```

---

## 📊 Benchmark: Tracing Overhead

| Operation      | Tracing Off | Tracing On | Overhead    |
|----------------|-------------|------------|-------------|
| HTTP Request   | ~1.201 ms   | ~1.248 ms  | +0.047 ms   |
| gRPC Call      | ~0.842 ms   | ~0.875 ms  | +0.033 ms   |
| Kafka Publish  | ~2.110 ms   | ~2.132 ms  | +0.022 ms   |

---

## 📁 Proje Yapısı

```
go-otel-kit/
├── docker-compose.yml          # Geliştirme ortamı (temel servisler)
├── docker-compose.full.yml     # Tam observability stack (tek komutla başlatılır)
├── otel-collector-config.yml   # OTel Collector pipeline konfigürasyonu
├── setup/                      # TracerProvider başlatma ve kapatma
├── middleware/                 # HTTP trace middleware
├── interceptor/                # gRPC trace interceptor
├── kafka/                      # Kafka producer/consumer trace wrapper
├── grafana/                    # Grafana dashboard ve provisioning dosyaları
└── examples/                   # Çalıştırılabilir örnek uygulamalar
    ├── http/                   # HTTP trace örneği
    ├── grpc/                   # gRPC trace örneği
    └── redis-pg/               # Redis + PostgreSQL trace örneği
```
