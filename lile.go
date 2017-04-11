// Package lile provides helper methods to quickly run gRPC based servers
// that have default metrics and tracing support
package lile

import (
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/lile/pubsub"
	"github.com/mwitkow/go-grpc-middleware"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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
	pubsubProvider pubsub.Provider
	publishers     map[string]string
	subscriber     pubsub.Subscriber
	tracing        bool
	tracer         *opentracing.Tracer
}

// An Option sets options
type Option func(*options)

// Server is a grpc compatible server with extra options
type Server struct {
	opts options
	*grpc.Server
}

func defaultOptions() options {
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

// Name sets the name of the service
func Name(n string) Option {
	return func(o *options) {
		o.name = n
	}
}

// Port sets the gRPC port of the service
func Port(n string) Option {
	return func(o *options) {
		o.port = n
	}
}

// PrometheusEnabled sets wether prometheus metrics are enabled
func PrometheusEnabled(b bool) Option {
	return func(o *options) {
		o.prometheus = b
	}
}

// PrometheusPort sets the prometheus metrics http port
func PrometheusPort(p string) Option {
	return func(o *options) {
		o.prometheusPort = p
	}
}

// PrometheusAddr sets the url for prometheus metrics (i.e. /metrics)
func PrometheusAddr(a string) Option {
	return func(o *options) {
		o.prometheusAddr = a
	}
}

// AddUnaryInterceptor adds a unary interceptor to the gRPC server
func AddUnaryInterceptor(unint grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryInts = append(o.unaryInts, unint)
	}
}

// AddStreamInterceptor adds a stream interceptor to the gRPC server
func AddStreamInterceptor(sint grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.streamInts = append(o.streamInts, sint)
	}
}

// Tracer adds an opentracing compatible tracer to the gRPC server
func Tracer(t opentracing.Tracer) Option {
	return func(o *options) {
		o.tracer = &t
	}
}

// TracingEnabled sets whether the intercept gRPC calls for tracing
func TracingEnabled(e bool) Option {
	return func(o *options) {
		o.tracing = e
	}
}

// Implementation registers the server handler for gRPC calls
func Implementation(impl registerImplementation) Option {
	return func(o *options) {
		o.implementation = impl
	}
}

// PubSubProvider registers the client for publishers and subscriptions
func PubSubProvider(p pubsub.Provider) Option {
	return func(o *options) {
		o.pubsubProvider = p
	}
}

// Publishers is a map[string]string of RPC methods and their event names, e.g
//
// lile.Publishers(map[string]string{
// 	"Create": "account_service.created",
// })
//
func Publishers(pubs map[string]string) Option {
	return func(o *options) {
		o.publishers = pubs
	}
}

// Subscriber registers the subscriber for pubsub subscriptions
func Subscriber(sub pubsub.Subscriber) Option {
	return func(o *options) {
		o.subscriber = sub
	}
}

// NewServer creates a lile server (gRPC server compatible) with N options
func NewServer(opt ...Option) *Server {
	opts := defaultOptions()
	for _, o := range opt {
		o(&opts)
	}

	if opts.prometheus {
		AddUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor)(&opts)
		AddStreamInterceptor(grpc_prometheus.StreamServerInterceptor)(&opts)
	}

	if opts.tracing {
		if opts.tracer == nil {
			opts.tracer = tracerFromEnv(opts)
		}

		if opts.tracer != nil {
			AddUnaryInterceptor(
				otgrpc.OpenTracingServerInterceptor(*opts.tracer),
			)(&opts)
		}
	}

	if opts.pubsubProvider == nil {
		opts.pubsubProvider = PubSubProviderFromEnv(opts)
	}

	client := &pubsub.Client{Provider: opts.pubsubProvider}

	if opts.publishers != nil {
		if opts.pubsubProvider == nil {
			logrus.Warnf("lile pubsub: publishers specified but no Provider is set")
		} else {
			client.InterceptorMethods = opts.publishers
			AddUnaryInterceptor(
				pubsub.UnaryServerInterceptor(client),
			)(&opts)
		}
	}

	if opts.subscriber != nil {
		if opts.pubsubProvider == nil {
			logrus.Warnf("lile pubsub: subscriber specified but no Provider is set")
		}

		opts.subscriber.Setup(client)
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

// ListenAndServe creates a tcp socket and starts listening for connections.
// it is NOT tls encrypted
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

// NewTestServer is a helper function to create a gRPC server on a unix socket
// it returns the socket location and a func to call which starts the server
func NewTestServer(opt ...Option) (string, func()) {
	// Create a temp random unix socket
	uid := uuid.NewV1().String()
	skt := "/tmp/" + uid

	ln, err := net.Listen("unix", skt)
	if err != nil {
		panic(err)
	}

	ts := NewServer(opt...)

	return skt, func() {
		ts.Serve(ln)
	}
}

// TestConn is a connection that connects to a sockets based test connection
func TestConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithDialer(func(addr string, d time.Duration) (net.Conn, error) {
			return net.Dial("unix", addr)
		}),
		grpc.WithInsecure(),
		grpc.WithTimeout(1*time.Second),
		grpc.WithBlock(),
	)

	if err != nil {
		panic(err)
	}

	return conn
}
