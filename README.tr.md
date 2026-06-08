[English](README.md) | [Türkçe](README.tr.md)

# go-otel-kit

`go-otel-kit`, Go uygulamaları için OpenTelemetry tabanlı dağıtık izleme (distributed tracing) kütüphanesidir. HTTP, gRPC, Kafka, Redis ve PostgreSQL için tekrar kullanılabilir izleme bileşenleri sağlar ve Jaeger ile Grafana üzerinde görselleştirme desteği sunar.

---

## Genel Bakış

Bu proje, aşağıdaki senaryolar için hazır araçlar sunar:

- OpenTelemetry TracerProvider başlatma,
- HTTP sunucu ve istemci izleme,
- gRPC interceptor entegrasyonu,
- Kafka üzerinden trace context aktarımı,
- Redis ve GORM destekli PostgreSQL sorgularının izlenmesi.

---

## Gereksinimler

- Docker
- Go 1.21+
- Git

---

## Hızlı Başlangıç

Tam izleme yığını seslendirmek için:

```bash
docker compose -f docker-compose.full.yml up -d
```

Çalışma durumunu kontrol edin:

```bash
docker compose -f docker-compose.full.yml ps
```

Yığını durdurmak için:

```bash
docker compose -f docker-compose.full.yml down
```

---

## Portlar

| Servis         | Port  | Açıklama                          |
|----------------|-------|-----------------------------------|
| Jaeger UI      | 16686 | Trace görselleştirme arayüzü      |
| OTel Collector | 4317  | OTLP gRPC bağlantı noktası        |
| OTel Collector | 4318  | OTLP HTTP bağlantı noktası        |
| OTel Collector | 8888  | Collector metrikleri              |
| Grafana        | 3000  | Dashboard arayüzü                 |
| Kafka          | 9092 | Kafka broker                      |
| PostgreSQL     | 5432 | Veritabanı                        |
| Redis          | 6379 | Cache                             |

---

## Arayüzler

- Jaeger: http://localhost:16686
- Grafana: http://localhost:3000
  - Kullanıcı adı: `admin`
  - Parola: `admin`

---

## Örnek HTTP İzleme

Örnek HTTP servisini çalıştırın:

```bash
cd examples/http
go run main.go
```

Test isteği gönderin:

```bash
curl http://localhost:8080/hello
```

Jaeger üzerinden `example-http-service` servisini seçerek trace verilerini inceleyebilirsiniz.

---

## Trace Görselleştirmeleri

Aşağıda `go-otel-kit` tarafından toplanan ve Jaeger arayüzünde görüntülenen gerçek dünya izleme (trace) örneklerini görebilirsiniz:

![Jaeger Gateway Trace](assets/jaeger-gatewaytrace.png)
![Jaeger Gateway Overview](assets/jaeger-gateway.png)
![Jaeger Backend Overview](assets/jaeger-backend.png)

---

## Entegrasyon Rehberi

### Kurulum

```bash
go get github.com/Chimera-State/go-otel-kit
```

### TracerProvider Başlatma

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
        log.Fatalf("Tracer başlatılamadı: %v", err)
    }
    defer setup.Shutdown(ctx)

    // Uygulama kodu
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

### Kafka Inject ve Extract

```go
// ÜRETİCİ
func PublishEvent(ctx context.Context, client *kgo.Client) {
    record := &kgo.Record{Topic: "orders-topic", Value: []byte("new-order")}
    kafka.InjectToRecord(ctx, record)
    client.Produce(ctx, record, nil)
}

// TÜKETİCİ
func ConsumeEvent(client *kgo.Client) {
    ctx := context.Background()
    fetches := client.PollFetches(ctx)
    iter := fetches.RecordIter()
    for !iter.Done() {
        record := iter.Next()
        extractedCtx := kafka.ExtractFromRecord(ctx, record)
        tracer := otel.Tracer("kafka-consumer")
        _, span := tracer.Start(extractedCtx, "process-order-event")

        // iş mantığı

        span.End()
    }
}
```

### Redis ve PostgreSQL (GORM)

```go
// Redis
redisotel.InstrumentTracing(rdb)

// GORM
db.Use(tracing.NewPlugin())
```

---

## Benchmark Sonuçları

Aşağıdaki değerler, tracing aktifleştirildiğinde ortaya çıkan ek gecikme örneklerini göstermektedir.

| Operasyon Tipi | Tracing Kapalı | Tracing Açık | Overhead    |
|----------------|---------------|--------------|-------------|
| HTTP Request   | ~1.201 ms     | ~1.248 ms    | +0.047 ms   |
| gRPC Call      | ~0.842 ms     | ~0.875 ms    | +0.033 ms   |
| Kafka Publish  | ~2.110 ms     | ~2.132 ms    | +0.022 ms   |

---

## Proje Yapısı

```
go-otel-kit/
├── docker-compose.yml          # Geliştirme ortamı
├── docker-compose.full.yml     # Tam izleme yığını
├── otel-collector-config.yml   # OTel Collector yapılandırması
├── setup/                      # TracerProvider başlatma ve kapatma
├── middleware/                 # HTTP izleme middleware'i
├── interceptor/                # gRPC interceptor
├── kafka/                      # Kafka trace yardımcıları
├── grafana/                    # Grafana dashboard ve provisioning dosyaları
└── examples/                   # Örnek uygulamalar
    ├── http/                   # HTTP örneği
    ├── grpc/                   # gRPC örneği
    └── redis-pg/               # Redis + PostgreSQL örneği
```
