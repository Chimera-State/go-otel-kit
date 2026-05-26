## Veritabanı ve Redis Çağrıları Nasıl Trace Edilir?

Redis ve PostgreSQL gibi araçları trace etmek için kendi uygulamalarınıza resmi kütüphaneleri eklemeniz yeterlidir:

**Redis için:**
Uygulamanızda `go-redis` client'ını oluştururken şu hook'u ekleyin:
`redisotel.InstrumentTracing(rdb)`

**PostgreSQL (GORM) için:**
Veritabanı bağlantınızı açtıktan sonra şu plugin'i ekleyin:
`db.Use(tracing.NewPlugin())`