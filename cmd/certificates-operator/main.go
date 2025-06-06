package main

import (
	"context"

	"github.com/gruyaume/certificates-operator/internal/charm"
	"github.com/gruyaume/charm-libraries/tracing"
	"github.com/gruyaume/goops"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	serviceName            = "certificates"
	TracingIntegrationName = "tracing"
)

func main() {
	env := goops.ReadEnv()

	if env.HookName == "" {
		return
	}

	run(env.HookName)
}

func run(hook string) {
	ctx, tp := initTracing()
	defer shutdown(tp, ctx)

	tracer := otel.Tracer(serviceName)
	ctx, span := tracer.Start(ctx, hook)

	defer span.End()

	charm.HandleDefaultHook(ctx)
	charm.SetStatus(ctx)

	flush(tp, ctx)
}

func initTracing() (context.Context, *trace.TracerProvider) {
	ti := tracing.Integration{
		RelationName: TracingIntegrationName,
		ServiceName:  serviceName,
	}
	ti.PublishSupportedProtocols([]tracing.Protocol{tracing.GRPC})

	ctx := context.Background()

	tp, err := ti.InitTracer(ctx)
	if err != nil {
		goops.LogErrorf("could not initialize tracer: %v", err)
		return ctx, nil
	}

	return ctx, tp
}

func flush(tp *trace.TracerProvider, ctx context.Context) {
	if tp != nil {
		tp.ForceFlush(ctx)
	}
}

func shutdown(tp *trace.TracerProvider, ctx context.Context) {
	if tp == nil {
		return
	}

	if err := tp.Shutdown(ctx); err != nil {
		goops.LogErrorf("could not shutdown tracer: %v", err)
	}
}
