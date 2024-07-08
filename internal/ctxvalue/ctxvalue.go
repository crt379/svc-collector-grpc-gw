package ctxvalue

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type ContextK interface {
	LoggerContextK | GrpcMetaContextK | TraceContextK
}

type CTargetType interface {
	zap.Logger | metadata.MD | string
}

type TargetContext[T CTargetType, K ContextK] struct {
	_ T
	_ K
}

func (tc TargetContext[T, K]) NewContext(ctx context.Context, v *T) context.Context {
	var k K
	return context.WithValue(ctx, k, v)
}

func (tc TargetContext[T, K]) GetValue(ctx context.Context) (*T, bool) {
	var k K
	v, ok := ctx.Value(k).(*T)
	return v, ok
}

type LoggerContextK struct{}

type LoggerContext struct {
	TargetContext[zap.Logger, LoggerContextK]
}

func (tc LoggerContext) GetValue(ctx context.Context) (*zap.Logger, bool) {
	var k LoggerContextK
	v, ok := ctx.Value(k).(*zap.Logger)
	if !ok {
		return zap.L(), ok
	}
	return v, ok
}

type GrpcMetaContextK struct{}

type GrpcMetaContext struct {
	TargetContext[metadata.MD, GrpcMetaContextK]
}

type TraceContextK struct{}

type TraceContext struct {
	TargetContext[string, TraceContextK]
}
