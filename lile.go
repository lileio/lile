// Package lile provides helper methods to quickly create RPC based services
// that have metrics, tracing and pub/sub support
package lile

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/lile/fromenv"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var service = NewService("lile")

type registerImplementation func(s *grpc.Server)

// Service is a grpc compatible server with extra features
type Service struct {
	Name string
	Port string
	// Interceptors
	UnaryInts  []grpc.UnaryServerInterceptor
	StreamInts []grpc.StreamServerInterceptor
	// The RPC server implementation
	GRPCImplementation registerImplementation
}

func defaultOptions(n string) Service {
	return Service{
		Name:               n,
		Port:               ":8000",
		GRPCImplementation: func(s *grpc.Server) {},
	}
}

// NewService creates a lile service with N options
func NewService(name string) Service {
	return defaultOptions(name)
}

func GlobalService() *Service {
	return &service
}

func Name(n string) {
	service.Name = n
}

func Port(n string) {
	service.Port = n
}

// AddUnaryInterceptor adds a unary interceptor to the RPC server
func AddUnaryInterceptor(unint grpc.UnaryServerInterceptor) {
	service.UnaryInts = append(service.UnaryInts, unint)
}

// AddStreamInterceptor adds a stream interceptor to the RPC server
func AddStreamInterceptor(sint grpc.StreamServerInterceptor) {
	service.StreamInts = append(service.StreamInts, sint)
}

func Server(r registerImplementation) {
	service.GRPCImplementation = r
}

func Serve() error {
	lis, err := net.Listen("tcp", service.Port)
	if err != nil {
		return err
	}

	logrus.Infof("Serving gRPC on %s", service.Port)
	return CreateServer().Serve(lis)
}

func CreateServer() *grpc.Server {
	// Default interceptors, [prometheus, opentracing]
	AddUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor)
	AddStreamInterceptor(grpc_prometheus.StreamServerInterceptor)
	AddUnaryInterceptor(otgrpc.OpenTracingServerInterceptor(
		fromenv.Tracer(service.Name)))

	gs := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(service.UnaryInts...)),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(service.StreamInts...)),
	)

	service.GRPCImplementation(gs)

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(gs)
	http.Handle("/metrics", prometheus.Handler())
	logrus.Infof("Prometheus metrics at :9000/metrics")
	port := "9000"
	if p := os.Getenv("PROMETHEUS_PORT"); p != "" {
		port = p
	}
	go http.ListenAndServe(":"+port, nil)

	return gs
}

// NewTestServer is a helper function to create a gRPC server on a unix socket
// it returns the socket location and a func to call which starts the server
func NewTestServer(s *grpc.Server) (string, func()) {
	// Create a temp random unix socket
	uid := uuid.NewV1().String()
	skt := "/tmp/" + uid

	ln, err := net.Listen("unix", skt)
	if err != nil {
		panic(err)
	}

	return skt, func() {
		s.Serve(ln)
	}
}

// TestConn is a connection that connects to a socket based connection
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
