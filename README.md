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

## 🔧 Uygulama Entegrasyonu

### Uygulamandan Trace Göndermek

OTel Collector, `localhost:4317` (gRPC) veya `localhost:4318` (HTTP) üzerinden trace kabul eder.  
`setup.Init()` fonksiyonu ile tracing altyapısını başlatabilirsin:

```go
err := setup.Init(ctx,
    setup.WithServiceName("benim-servisim"),
    setup.WithServiceVersion("1.0.0"),
    // Varsayılan endpoint: localhost:4317
)
```

---

### Veritabanı ve Redis Çağrıları Nasıl Trace Edilir?

**Redis için:**  
`go-redis` client'ını oluştururken şu hook'u ekle:

```go
redisotel.InstrumentTracing(rdb)
```

**PostgreSQL (GORM) için:**  
Veritabanı bağlantısını açtıktan sonra şu plugin'i ekle:

```go
db.Use(tracing.NewPlugin())
```

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