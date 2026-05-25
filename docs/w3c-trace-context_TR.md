# W3C Trace Context ve Propagation Teknik Özeti

Uçtan uca izlenebilirlik (End-to-End Observability) sağlayabilmek için, bir HTTP isteğinin veya Kafka mesajının servisler arasında dolaşırken "kimliğini" kaybetmemesi gerekir. OpenTelemetry, bu kimliği servisler arasında taşımak için endüstri standardı olan **W3C Trace Context** formatını kullanır.

## 1. `traceparent` Header Formatı Nedir?
Bir servis diğerine istek attığında (örneğin A servisi B servisine HTTP isteği yaparken), HTTP Header'larına (veya Kafka Header'larına) `traceparent` adında tek bir satır eklenir. 

Bu header'ın formatı şu şekildedir:
**`traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01`**

Bu 4 bölümlü yapı şu anlama gelir:
1. **`version` (00):** W3C formatının versiyonudur. Şu an sabit olarak `00` kullanılır.
2. **`trace-id` (0af7651916cd43dd8448eb211c80319c):** İsteğin en başından en sonuna kadar (tüm servisler boyunca) **asla değişmeyen** 16-byte'lık benzersiz kimliktir. Bütün span'ler bu şemsiye altında toplanır.
3. **`parent-id` (b7ad6b7169203331):** İsteği gönderen *bir önceki span'in (işlemin)* 8-byte'lık kimliğidir. B servisi bu id'yi okur ve "Benim parent'ım buymuş" diyerek zincire eklenir. (Aynı zamanda `span-id` olarak da bilinir).
4. **`trace-flags` (01):** Bu izlemenin örneklenip örneklenmediğini (sampled) belirtir. `01` ise bu isteğin metrikleri Collector/Jaeger'a gönderilecek demektir, `00` ise gönderilmeyecek (drop edilecek) demektir.

## 2. Inject Kavramı
Bir servis (örneğin Client), başka bir servise istek atmadan *hemen önce* kendi içindeki mevcut Trace Context'ini alıp, HTTP isteğinin (veya Kafka mesajının) Header'larına **yazması/gömmesi** işlemidir.
- **Ne zaman yapılır?** Dışarıya bir istek veya mesaj çıkarken (HTTP Client, gRPC Client, Kafka Producer).
- *Özetle: "Giden kargoya takip numarasını (traceparent) yapıştırmak."*

## 3. Extract Kavramı
Bir servis (örneğin Server), dışarıdan bir istek aldığında, gelen isteğin Header'larındaki `traceparent` bilgisini **okuması** ve kendi içindeki Context'e (Go'daki `context.Context`) aktarması işlemidir. Eğer gelen istekte bu header yoksa, yeni ve yepyeni bir trace başlatılır.
- **Ne zaman yapılır?** Dışarıdan bir istek veya mesaj içeri girerken (HTTP Server Handler, gRPC Server Interceptor, Kafka Consumer).
- *Özetle: "Gelen kargonun üzerindeki takip numarasını okuyup sisteme kaydetmek."*
