package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	defaultPrometheusEnabled = true
)

type Server struct {
	host       string
	port       int
	listener   net.Listener
	handler    RegisterGRPCHandlerFunc
	options    Options
	server     *grpc.Server
	httpServer *http.Server
}

type RegisterGRPCHandlerFunc func(server *grpc.Server)

type RegisterHTTPHandlerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

func NewServer(host string, port int, handler RegisterGRPCHandlerFunc, opts ...Option) *Server {
	s := &Server{
		host: host,
		port: port,
		options: Options{
			isPrometheusEnabled: defaultPrometheusEnabled,
		},
		handler: handler,
	}

	// apply options
	for _, o := range opts {
		o(&s.options)
	}

	return s
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}

	s.listener = lis

	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigs

		shutdownErr := s.Stop()
		if shutdownErr != nil {
			fmt.Println("Shutdown error", zap.Error(shutdownErr))
		}

		cancel()
		s.runHook()
		done <- shutdownErr
	}()

	go s.serve(ctx)

	fmt.Println("Server now listening")
	fmt.Println("Ctrl-C to interrupt...")
	e := <-done
	fmt.Println("Exiting....", zap.Error(e))
	return e
}

// start listening grpc & http request
func (s *Server) serve(ctx context.Context) {
	if s.options.httpOptions.registerHandler != nil {
		m := cmux.New(s.listener)
		grpcListener := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
		httpListener := m.Match(cmux.HTTP1Fast())

		g := new(errgroup.Group)
		g.Go(func() error { return s.grpcServe(ctx, grpcListener) })
		g.Go(func() error { return s.httpServe(ctx, httpListener) })
		g.Go(func() error { return m.Serve() })

		g.Wait()
	} else {
		s.grpcServe(ctx, s.listener)
	}
}

func (s *Server) grpcServe(ctx context.Context, l net.Listener) error {
	sIntOpt := s.configureStreamInterceptors()

	uIntOpt := s.configureUnaryInterceptors()

	server := grpc.NewServer(sIntOpt, uIntOpt)

	if s.handler != nil {
		s.handler(server)
	}

	reflection.Register(server)

	if s.options.isPrometheusEnabled {
		grpc_prometheus.Register(server)
		grpc_prometheus.EnableHandlingTimeHistogram()
	}

	s.server = server
	return server.Serve(l)
}

func (s *Server) configureStreamInterceptors() grpc.ServerOption {
	chain := make([]grpc.StreamServerInterceptor, 0, 10)
	// auth
	if s.options.grpcOptions.authInterceptor != nil {
		chain = append(chain, s.options.grpcOptions.authInterceptor)
	}

	// tracing
	chain = append(chain,
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(s.options.tracer)),
	)

	// prometheus
	if s.options.isPrometheusEnabled {
		chain = append(chain, grpc_prometheus.StreamServerInterceptor)
	}

	// prevent panic
	chain = append(chain, grpc_recovery.StreamServerInterceptor())

	// middlewares from caller
	if len(s.options.grpcOptions.streamInterceptors) > 0 {
		chain = append(chain, s.options.grpcOptions.streamInterceptors...)
	}

	return grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(chain...))
}

func (s *Server) configureUnaryInterceptors() grpc.ServerOption {
	chain := make([]grpc.UnaryServerInterceptor, 0, 10)
	// auth
	if s.options.httpOptions.authInterceptor != nil {
		chain = append(chain, s.options.httpOptions.authInterceptor)
	}

	// tracing
	chain = append(chain,
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(s.options.tracer)),
	)

	// prometheus
	if s.options.isPrometheusEnabled {
		chain = append(chain, grpc_prometheus.UnaryServerInterceptor)
	}

	// prevent panic
	chain = append(chain, grpc_recovery.UnaryServerInterceptor())

	// middlewares from caller
	if len(s.options.grpcOptions.unaryInterceptors) > 0 {
		chain = append(chain, s.options.grpcOptions.unaryInterceptors...)
	}

	return grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(chain...))
}

// serve http request
func (s *Server) httpServe(ctx context.Context, l net.Listener) error {
	// configure mux options
	muxOpts := []runtime.ServeMuxOption{
		runtime.WithMarshalerOption("*", &runtime.JSONPb{
			OrigName:     true,
			EnumsAsInts:  true,
			EmitDefaults: true,
		}),
	}
	if s.options.httpOptions.serverMuxOpts != nil {
		muxOpts = append(muxOpts, s.options.httpOptions.serverMuxOpts...)
	}
	mux := runtime.NewServeMux(muxOpts...)

	// Register handler
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	opts := []grpc.DialOption{creds}
	endPoint := fmt.Sprintf("localhost:%d", s.port)

	err := s.options.httpOptions.registerHandler(ctx, mux, endPoint, opts)
	if err != nil {
		return err
	}

	// add handler middlewares
	var handler http.Handler
	handler = mux

	// chain middleware functions
	for _, interceptor := range s.options.httpOptions.interceptors {
		handler = interceptor(handler)
	}

	server := &http.Server{Handler: handler}
	s.httpServer = server

	return server.Serve(l)
}

// Stop shutdown this server gracefully
func (s *Server) Stop() error {
	// http
	s.gracefulShutdownHTTP()

	// grpc
	s.gracefulShutdownGRPC()

	fmt.Println("Shutting down server")
	return nil
}

func (s *Server) runHook() {
	for _, hook := range s.options.shutdownHooks {
		defer hook()
	}
}

// shutdown http server gracefully
func (s *Server) gracefulShutdownHTTP() {
	if s.httpServer == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		fmt.Println("Shutdown http server error", zap.Error(err))
	}
}

// shutdown grpc server gracefully
func (s *Server) gracefulShutdownGRPC() {
	if s.server == nil {
		return
	}
	s.server.Stop()
}
