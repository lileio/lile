package lile

import (
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/mwitkow/go-grpc-middleware"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
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
	tracing        bool
	tracer         *opentracing.Tracer
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
		tracing:        true,
		tracer:         nil,
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

func Tracer(t opentracing.Tracer) Option {
	return func(o *options) {
		o.tracer = &t
	}
}

func TracingEnabled(e bool) Option {
	return func(o *options) {
		o.tracing = e
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

	if opts.tracing {
		if opts.tracer == nil {
			t := tracerFromEnv(opts)
			opts.tracer = &t
		}

		AddUnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(*opts.tracer),
		)(&opts)
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

	logrus.Infof("Serving %s: gRPC %s", s.opts.name, s.opts.port)
	if s.opts.prometheus {
		logrus.Infof("Prometeus metrics on %s %s", s.opts.prometheusAddr, s.opts.prometheusPort)
	}

	return s.Serve(lis)
}
