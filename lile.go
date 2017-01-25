package lile

import (
	"net"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc"
)

type registerImplementation func(s *grpc.Server)

type options struct {
	name           string
	port           string
	prometheus     bool
	prometheusPort string
	prometheusAddr string
	unaryInts      []grpc.UnaryServerInterceptor
	streamInts     []grpc.StreamServerInterceptor
	implementation registerImplementation
}

type Option func(*options)

type Server struct {
	opts options
	*grpc.Server
}

func DefaultOptions() options {
	return options{
		name:           "lile_service",
		port:           ":8000",
		prometheus:     true,
		prometheusPort: ":8080",
		prometheusAddr: "/metrics",
		implementation: func(s *grpc.Server) {},
	}
}

func Name(n string) Option {
	return func(o *options) {
		o.name = n
	}
}

func AddUnaryInterceptor(unint grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryInts = append(o.unaryInts, unint)
	}
}

func AddStreamInterceptor(sint grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.streamInts = append(o.streamInts, sint)
	}
}

func Implementation(impl registerImplementation) Option {
	return func(o *options) {
		o.implementation = impl
	}
}

func NewServer(opt ...Option) *Server {
	opts := DefaultOptions()
	for _, o := range opt {
		o(&opts)
	}

	if opts.prometheus {
		AddUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor)(&opts)
		AddStreamInterceptor(grpc_prometheus.StreamServerInterceptor)(&opts)
	}

	s := grpc.NewServer(
		// Interceptors
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(opts.unaryInts...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(opts.streamInts...)),
	)

	opts.implementation(s)

	if opts.prometheus {
		grpc_prometheus.Register(s)

		mux := http.NewServeMux()
		mux.Handle(opts.prometheusAddr, prometheus.Handler())
		go http.ListenAndServe(opts.prometheusPort, mux)
	}

	return &Server{opts, s}
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.opts.port)
	if err != nil {
		return err
	}

	log.Infof("Serving %s: gRPC %s", s.opts.name, s.opts.port)
	if s.opts.prometheus {
		log.Infof("Prometheus metrics on %s %s", s.opts.prometheusAddr, s.opts.prometheusPort)
	}

	return s.Serve(lis)
}

// func KafkaTracer(addr string) (t opentracing.Tracer, err error) {
// 	collector, err := zipkin.NewKafkaCollector([]string{addr})
// 	if err != nil {
// 		return t, err
// 	}

// 	t, err = zipkin.NewTracer(
// 		zipkin.NewRecorder(collector, false, "account_service", "Account Service"),
// 		zipkin.ClientServerSameSpan(true), // for Zipkin V1 RPC span style
// 	)
// 	if err != nil {
// 		return t, err
// 	}

// 	return t, err
// }
