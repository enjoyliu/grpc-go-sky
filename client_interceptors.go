// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package grpc_go_sky

import (
	"context"
	"strings"

	"github.com/SkyAPM/go2sky"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

// UnaryClientInterceptor returns a new unary server interceptors.
func UnaryClientInterceptor(tracer *go2sky.Tracer, opts ...Option) grpc.UnaryClientInterceptor {
	options := &options{
		reportTags: []string{},
	}
	for _, o := range opts {
		o(options)
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
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

			if md, ok := metadata.FromOutgoingContext(ctx); ok {
				for _, k := range options.reportTags {
					span.Tag(go2sky.Tag(k), strings.Join(md.Get(k), ""))
				}
			}
			return err
		}
		return nil
	}

}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(tracer *go2sky.Tracer, opts ...Option) grpc.StreamClientInterceptor {
	options := &options{
		reportTags: []string{},
	}
	for _, o := range opts {
		o(options)
	}

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			span, ctx, err := tracer.CreateEntrySpan(ctx, "", func(key string) (string, error) {
				return strings.Join(md.Get(key), ""), nil
			})
			if err != nil {
				return clientStream, err
			}
			defer func() { span.End() }()

			span.SetComponent(componentIDGrpcGo)
			span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

			if md, ok := metadata.FromOutgoingContext(ctx); ok {
				for _, k := range options.reportTags {
					span.Tag(go2sky.Tag(k), strings.Join(md.Get(k), ""))
				}
			}
			return clientStream, err
		}
		return clientStream, err
	}
}
