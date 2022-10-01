package rpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

const (
	defaultRequestTimeout    = 30 * time.Second
	defaultDialTimeout       = 15 * time.Second
	defaultPrometheusEnabled = true
)

func NewClient(target string, opts ...Option) (*grpc.ClientConn, error) {
	options := configure(opts...)
	dialOpts := options.dialOptions

	dialOpts = append(dialOpts,
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
	)

	sIntOpt := grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
		grpc_opentracing.StreamClientInterceptor(grpc_opentracing.WithTracer(options.tracer)),
		grpc_prometheus.StreamClientInterceptor,
		grpc_retry.StreamClientInterceptor(options.retryPolicies...),
	))

	dialOpts = append(dialOpts, sIntOpt)

	if options.isPrometheusEnabled {
		grpc_prometheus.EnableClientHandlingTimeHistogram()
	}

	uIntOpt := grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
		grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(options.tracer)),
		grpc_prometheus.UnaryClientInterceptor,
		grpc_retry.UnaryClientInterceptor(options.retryPolicies...),
	))

	dialOpts = append(dialOpts, uIntOpt)

	ctx, cancel := context.WithTimeout(context.Background(), defaultDialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, dialOpts...)

	return conn, err
}

// apply options
func configure(opts ...Option) *Options {
	options := &Options{
		dialTimeout: defaultDialTimeout,
		retryPolicies: []grpc_retry.CallOption{
			grpc_retry.WithBackoff(
				grpc_retry.BackoffExponential(50 * time.Millisecond),
			),
			grpc_retry.WithCodes(codes.Unavailable),
			grpc_retry.WithMax(3),
		},
		dialOptions:         make([]grpc.DialOption, 0, 10),
		isPrometheusEnabled: defaultPrometheusEnabled,
	}

	// apply all options
	for _, o := range opts {
		o(options)
	}

	if options.requestTimeout == 0 {
		options.requestTimeout = defaultRequestTimeout
	}

	options.retryPolicies = append(options.retryPolicies, grpc_retry.WithPerRetryTimeout(options.requestTimeout))

	return options
}
