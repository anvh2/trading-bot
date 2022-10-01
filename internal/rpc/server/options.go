package rpc

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

type ShutdownHook func()

// Options configure how to setup gRPC server
type Options struct {
	grpcOptions         grpcOptions
	httpOptions         httpOptions
	isPrometheusEnabled bool
	shutdownHooks       []ShutdownHook
	tracer              opentracing.Tracer
}

// Option wraps a function that modifies Options.
type Option func(*Options)

type grpcOptions struct {
	authInterceptor    grpc.StreamServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
}

type httpOptions struct {
	authInterceptor grpc.UnaryServerInterceptor
	registerHandler RegisterHTTPHandlerFunc
	interceptors    []HTTPInterceptor
	serverMuxOpts   []runtime.ServeMuxOption
}

// HTTPInterceptor middleware for http handler
type HTTPInterceptor func(handler http.Handler) http.Handler

// StreamInterceptor add interceptor while proceed stream RPC request
func StreamInterceptor(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		o.grpcOptions.streamInterceptors = interceptors
	}
}

// StreamAuth add authentication mechanism before proceed request
func StreamAuth(interceptor grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		o.grpcOptions.authInterceptor = interceptor
	}
}

// UnaryInterceptor add interceptor while proceed unary RPC request
func UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.grpcOptions.unaryInterceptors = interceptors
	}
}

// UnaryAuth add authentication mechanism before proceed request
func UnaryAuth(interceptor grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.httpOptions.authInterceptor = interceptor
	}
}

// RegisterHTTPHandler invoke RegisterServiceHandlerFromEndpoint method
func RegisterHTTPHandler(registerHandlerFunc RegisterHTTPHandlerFunc) Option {
	return func(o *Options) {
		o.httpOptions.registerHandler = registerHandlerFunc
	}
}

// MuxInterceptor add middleware for http handler
func MuxInterceptor(interceptors ...HTTPInterceptor) Option {
	return func(o *Options) {
		if o.httpOptions.interceptors == nil {
			o.httpOptions.interceptors = interceptors
		} else {
			o.httpOptions.interceptors = append(o.httpOptions.interceptors, interceptors...)
		}
	}
}

// MuxOption add middleware into muxing connection
func MuxOption(muxOptions ...runtime.ServeMuxOption) Option {
	return func(o *Options) {
		if o.httpOptions.serverMuxOpts == nil {
			o.httpOptions.serverMuxOpts = muxOptions
		} else {
			o.httpOptions.serverMuxOpts = append(o.httpOptions.serverMuxOpts, muxOptions...)
		}
	}
}

// EnablePrometheus allow prometheus tracking when proceed request. Default is true
func EnablePrometheus(enabled bool) Option {
	return func(o *Options) {
		o.isPrometheusEnabled = enabled
	}
}

// WithShutdownHook hook to shutdown process
func WithShutdownHook(hooks ...ShutdownHook) Option {
	return func(o *Options) {
		if o.shutdownHooks == nil {
			o.shutdownHooks = hooks
		} else {
			o.shutdownHooks = append(o.shutdownHooks, hooks...)
		}
	}
}

// WithTracer add tracer for tracing request
func WithTracer(t opentracing.Tracer) Option {
	return func(o *Options) {
		o.tracer = t
	}
}
