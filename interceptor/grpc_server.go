package interceptor

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const tracerName = "github.com/Chimera-State/go-otel-kit/interceptor"

type metadataCarrier metadata.MD

func (mc metadataCarrier) Get(key string) string {
	if values := metadata.MD(mc).Get(key); len(values) > 0 {
		return values[0]
	}
	return ""
}

func (mc metadataCarrier) Set(key string, value string) {
	metadata.MD(mc).Set(key, value)
}

func (mc metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range mc {
		keys = append(keys, k)
	}
	return keys
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		ctx = otel.GetTextMapPropagator().Extract(ctx, metadataCarrier(md))

		//yeni span başlat
		tracer := otel.Tracer(tracerName)
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		//gRPC işlemini çalıştır
		resp, err := handler(ctx, req)

		//hata ise Error de
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		return resp, err
	}
}

func ServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(UnaryServerInterceptor()),
	}
}
