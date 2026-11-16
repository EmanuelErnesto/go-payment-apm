package logger

import (
	"context"
	"io"

	"go.elastic.co/apm/module/apmzap/v2"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Options is defined by the adapter itself and filled in by the composition
// root (cmd/), so this package never reads the environment.
type Options struct {
	Level       zapcore.Level
	ServiceName string
	Environment string
	// Output is where ECS JSON lines are written. In every environment this is
	// stdout: shipping logs is the job of the platform (Filebeat -> Logstash),
	// not of the application.
	Output io.Writer
}

// New builds a zap logger whose encoder is Elastic's official ECS encoder
// (ecszap), so @timestamp, log.level, message, error.* and log.origin already
// follow the Elastic Common Schema and need no remapping in Logstash.
func New(opts Options) *zap.Logger {
	core := ecszap.NewCore(
		ecszap.NewDefaultEncoderConfig(),
		zapcore.Lock(zapcore.AddSync(opts.Output)),
		opts.Level,
	)

	return zap.New(core, zap.AddCaller()).With(
		zap.String("service.name", opts.ServiceName),
		zap.String("service.environment", opts.Environment),
	)
}

type ctxKey struct{}

// WithContext stores a logger in ctx. The HTTP middleware enriches the base
// logger with the APM trace ids (trace.id, transaction.id, span.id) once per
// request, so any layer that receives the context logs already correlated.
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext returns the request-scoped logger, falling back to the global
// one (set with zap.ReplaceGlobals in cmd/) outside of a request.
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	}
	return zap.L()
}

// WithTrace returns a child logger carrying the APM correlation ids found in
// ctx, using Elastic's official apmzap module.
func WithTrace(ctx context.Context, l *zap.Logger) *zap.Logger {
	return l.With(apmzap.TraceContext(ctx)...)
}
