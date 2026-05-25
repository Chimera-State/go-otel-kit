# W3C Trace Context and Propagation Technical Summary

To achieve End-to-End Observability, an HTTP request or Kafka message must not lose its "identity" as it travels across services. OpenTelemetry uses the **W3C Trace Context** format, an industry standard, to propagate this identity between services.

## 1. What is the `traceparent` Header Format?
When one service makes a request to another (e.g., Service A making an HTTP request to Service B), a single line named `traceparent` is added to the HTTP Headers (or Kafka Headers). 

The format of this header is as follows:
**`traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01`**

This 4-part structure means the following:
1. **`version` (00):** The version of the W3C format. Currently, `00` is used as a constant.
2. **`trace-id` (0af7651916cd43dd8448eb211c80319c):** A 16-byte unique identifier that **never changes** from the very beginning to the end of the request (across all services). All spans are grouped under this umbrella.
3. **`parent-id` (b7ad6b7169203331):** The 8-byte identifier of the *previous span (operation)* that sent the request. Service B reads this ID and attaches itself to the chain, effectively saying "This is my parent." (Also known as `span-id`).
4. **`trace-flags` (01):** Indicates whether this trace is sampled or not. `01` means the metrics for this request will be sent to the Collector/Jaeger, while `00` means they will not be sent (dropped).

## 2. The Concept of Inject
It is the process where a service (e.g., Client) takes its current Trace Context and **writes/embeds** it into the Headers of the HTTP request (or Kafka message) *just before* making a request to another service.
- **When is it done?** When a request or message is going out (HTTP Client, gRPC Client, Kafka Producer).
- *In short: "Pasting the tracking number (traceparent) onto the outgoing package."*

## 3. The Concept of Extract
It is the process where a service (e.g., Server), upon receiving an external request, **reads** the `traceparent` information from the incoming request's Headers and transfers it into its internal Context (`context.Context` in Go). If this header is missing from the incoming request, a brand new trace is started.
- **When is it done?** When a request or message is coming in (HTTP Server Handler, gRPC Server Interceptor, Kafka Consumer).
- *In short: "Reading the tracking number on the incoming package and registering it into the system."*