package rpc

import (
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Options configure how to setup gRPC client
type Options struct {
	tracer              opentracing.Tracer
	dialTimeout         time.Duration
	dialOptions         []grpc.DialOption
	retryPolicies       []grpc_retry.CallOption
	requestTimeout      time.Duration
	isPrometheusEnabled bool
}

// Option wraps a function that modifies Options.
type Option func(*Options)

// DialTimeout timeout when creating grpc connection to server
func DialTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.dialTimeout = timeout
	}
}

// RequestTimeout timeout of each request.
// If called, it will set the request_timeout in retry_pilicies
// otherwise it will use default 30s per request
// Total_timeout = request_timeout x max_retry
func RequestTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.requestTimeout = timeout
	}
}

// RetryPolicies custom retry configuaration,
// default configuartion is:
//    - max_retry: 3
//    - back_off: 50ms
//    - timeout: 30s
//    - codes: 14-unavailable
func RetryPolicies(retryPolicies ...grpc_retry.CallOption) Option {
	return func(opts *Options) {
		opts.retryPolicies = retryPolicies
	}
}

// Prometheus allow prometheus tracking when proceed request. Default is true
func Prometheus(enabled bool) Option {
	return func(o *Options) {
		o.isPrometheusEnabled = enabled
	}
}

// DialOption add more gRPC dial options
func DialOption(option ...grpc.DialOption) Option {
	return func(o *Options) {
		o.dialOptions = append(o.dialOptions, option...)
	}
}

// WithInsecure open insecure connection
func WithInsecure() Option {
	return func(o *Options) {
		creds := grpc.WithTransportCredentials(insecure.NewCredentials())
		o.dialOptions = append(o.dialOptions, creds)
	}
}

// WithTransportCredentials configure connection credentials (TLS/SSL)
func WithTransportCredentials(creds credentials.TransportCredentials) Option {
	return func(o *Options) {
		o.dialOptions = append(o.dialOptions, grpc.WithTransportCredentials(creds))
	}
}

// WithBlock open insecure connection
func WithBlock() Option {
	return func(o *Options) {
		o.dialOptions = append(o.dialOptions, grpc.WithBlock())
	}
}

// WithTracer add tracer for tracing request
func WithTracer(t opentracing.Tracer) Option {
	return func(o *Options) {
		o.tracer = t
	}
}
