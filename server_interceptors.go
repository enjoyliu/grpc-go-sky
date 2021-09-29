// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package grpc_go_sky

import (
	"context"
	"fmt"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/SkyAPM/go2sky"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	componentIDGrpcGo = 6666
)

type Option func(*options)

type options struct {
	reportTags []string
}

// WithReportTags will set tags that need to report in metadata
func WithReportTags(tags ...string) Option {
	return func(o *options) {
		o.reportTags = append(o.reportTags, tags...)
	}
}

// UnaryServerInterceptor returns a new unary server interceptors .
func UnaryServerInterceptor(tracer *go2sky.Tracer, opts ...Option) grpc.UnaryServerInterceptor {
	options := &options{
		reportTags: []string{},
	}
	for _, o := range opts {
		o(options)
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			span, ctx, err := tracer.CreateEntrySpan(ctx, "", func(key string) (string, error) {
				return strings.Join(md.Get(key), ""), nil
			})
			if err != nil {
				return nil, err
			}
			defer func() { span.End() }()

			span.SetComponent(componentIDGrpcGo)
			span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

			if md, ok := metadata.FromIncomingContext(ctx); ok {
				for _, k := range options.reportTags {
					span.Tag(go2sky.Tag(k), strings.Join(md.Get(k), ""))
				}
			}
			fmt.Printf("%+v, %+v", ctx, req)
			reply, err := handler(ctx, req)
			if err != nil {
				span.Error(time.Now(), err.Error())
			}
			return reply, err
		} else {
			fmt.Printf("%+v, %+v", ctx, req)
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor.
func StreamServerInterceptor(tracer *go2sky.Tracer, opts ...Option) grpc.StreamServerInterceptor {
	options := &options{
		reportTags: []string{},
	}
	for _, o := range opts {
		o(options)
	}
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = stream.Context()
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			span, ctx, err := tracer.CreateEntrySpan(ctx, "", func(key string) (string, error) {
				return strings.Join(md.Get(key), ""), nil
			})
			if err != nil {
				return err
			}
			defer func() { span.End() }()

			span.SetComponent(componentIDGrpcGo)
			span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

			if md, ok := metadata.FromIncomingContext(ctx); ok {
				for _, k := range options.reportTags {
					span.Tag(go2sky.Tag(k), strings.Join(md.Get(k), ""))
				}
			}

			fmt.Printf("%+v, %+v", ctx, wrapped)
			err = handler(ctx, wrapped)
			if err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		} else {
			fmt.Printf("%+v, %+v", ctx, wrapped)
		}
		return handler(ctx, wrapped)
	}
}
