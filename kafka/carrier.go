package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
)

type MessageCarrier struct {
	record *kgo.Record
}

func NewCarrier(record *kgo.Record) MessageCarrier {
	return MessageCarrier{record: record}
}

func (c MessageCarrier) Get(key string) string {
	for _, h := range c.record.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c MessageCarrier) Set(key, value string) {
	c.record.Headers = append(c.record.Headers, kgo.RecordHeader{
		Key:   key,
		Value: []byte(value),
	})
}

func (c MessageCarrier) Keys() []string {
	keys := make([]string, len(c.record.Headers))
	for i, h := range c.record.Headers {
		keys[i] = h.Key
	}
	return keys
}

func InjectToRecord(ctx context.Context, record *kgo.Record) {
	otel.GetTextMapPropagator().Inject(ctx, NewCarrier(record))
}

func ExtractFromRecord(ctx context.Context, record *kgo.Record) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, NewCarrier(record))
}
