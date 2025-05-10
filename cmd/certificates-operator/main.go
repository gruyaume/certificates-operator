package main

import (
	"context"

	"github.com/gruyaume/certificates-operator/internal/charm"
	"github.com/gruyaume/goops"
	"github.com/gruyaume/goops/commands"
	"github.com/gruyaume/notary-k8s-operator/integrations/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	serviceName            = "certificates"
	TracingIntegrationName = "tracing"
)

func main() {
	hc := goops.NewHookContext()
	hook := hc.Environment.JujuHookName()

	if hook == "" {
		return
	}

	run(hc, hook)
}

func run(hc *goops.HookContext, hook string) {
	ctx, tp := initTracing(hc)
	defer shutdown(tp, ctx)

	tracer := otel.Tracer(serviceName)
	ctx, span := tracer.Start(ctx, hook)

	defer span.End()

	charm.HandleDefaultHook(ctx, hc)
	charm.SetStatus(ctx, hc)

	flush(tp, ctx)
}

func initTracing(hc *goops.HookContext) (context.Context, *trace.TracerProvider) {
	ti := tracing.Integration{
		HookContext:  hc,
		RelationName: TracingIntegrationName,
		ServiceName:  serviceName,
	}
	ti.PublishSupportedProtocols([]tracing.Protocol{tracing.GRPC})

	ctx := context.Background()

	tp, err := ti.InitTracer(ctx)
	if err != nil {
		hc.Commands.JujuLog(commands.Error, "could not initialize tracer:", err.Error())
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
		hc := goops.NewHookContext()
		hc.Commands.JujuLog(commands.Error, "could not shutdown tracer:", err.Error())
	}
}
