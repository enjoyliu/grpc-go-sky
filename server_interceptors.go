// Copyright (c) Improbable Worlds Ltd, All Rights Reserved

package grpc_go_sky

import (
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"strings"
	"time"
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


// UnaryServerInterceptor returns a new unary server interceptors that adds logrus.Entry to the context.
func UnaryServerInterceptor(tracer *go2sky.Tracer, opts ...Option) grpc.UnaryServerInterceptor {
	options := &options{
		reportTags: []string{},
	}
	for _, o := range opts {
		o(options)
	}
		return func (ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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


