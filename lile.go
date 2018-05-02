// Package lile provides helper methods to quickly create RPC based services
// that have metrics, tracing and pub/sub support
package lile

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/fromenv"
	"github.com/mattn/go-colorable"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	service = NewService("lile")
	once    sync.Once
)

type RegisterImplementation func(s *grpc.Server)

// ServerConfig is a generic server configuration
type ServerConfig struct {
	Port int
	Host string
}

func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Service is a gRPC based server with extra features
type Service struct {
	ID   string
	Name string

	// Interceptors
	UnaryInts  []grpc.UnaryServerInterceptor
	StreamInts []grpc.StreamServerInterceptor

	// The RPC server implementation
	GRPCImplementation RegisterImplementation
	GRPCOptions        []grpc.ServerOption

	// gRPC and Prometheus endpoints
	Config           ServerConfig
	PrometheusConfig ServerConfig

	// Registery allows Lile to work with external registeries like
	// consul, zookeeper or similar
	Registery Registery

	// Private utils, exposed so they can be useful if needed
	ServiceListener  net.Listener
	GRPCServer       *grpc.Server
	PrometheusServer *http.Server
}

// NewService creates a new service with a given name
func NewService(n string) *Service {
	once.Do(func() {
		if runtime.GOOS == "windows" {
			logrus.SetOutput(colorable.NewColorableStdout())
			logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
		}
	})
	return &Service{
		ID:                 generateID(n),
		Name:               n,
		Config:             ServerConfig{Host: "0.0.0.0", Port: 8000},
		PrometheusConfig:   ServerConfig{Host: "0.0.0.0", Port: 9000},
		GRPCImplementation: func(s *grpc.Server) {},
		UnaryInts: []grpc.UnaryServerInterceptor{
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
			otgrpc.OpenTracingServerInterceptor(
				fromenv.Tracer(n)),
		},
		StreamInts: []grpc.StreamServerInterceptor{
			grpc_prometheus.StreamServerInterceptor,
			grpc_recovery.StreamServerInterceptor(),
		},
	}
}

// GlobalService returns the global service
func GlobalService() *Service {
	return service
}

// Name sets the name for the service
func Name(n string) {
	service.ID = generateID(n)
	service.Name = n
}

// Server attaches the gRPC implementation to the service
func Server(r func(s *grpc.Server)) {
	service.GRPCImplementation = r
}

// AddUnaryInterceptor adds a unary interceptor to the RPC server
func AddUnaryInterceptor(unint grpc.UnaryServerInterceptor) {
	service.UnaryInts = append(service.UnaryInts, unint)
}

// AddStreamInterceptor adds a stream interceptor to the RPC server
func AddStreamInterceptor(sint grpc.StreamServerInterceptor) {
	service.StreamInts = append(service.StreamInts, sint)
}

// URLForService returns a service URL via a registry or a simple DNS name
// if not available via the registery
func URLForService(name string) string {
	if service.Registery != nil {
		registeryURL, err := service.Registery.Get(name)
		if err != nil {
			fmt.Printf("lile: error contacting registery for service %s. err: %s \n", name, err.Error())
		}

		return registeryURL
	}

	return name + ":80"
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
