// Package lile provides helper methods to quickly create RPC based services
// that have metrics, tracing and pub/sub support
package lile

import (
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/lile/fromenv"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	service        = NewService("lile")
	grpcHost       string
	grpcPort       int
	prometheusHost string
	prometheusPort int

	defaultHost           = "0.0.0.0"
	defaultGRPCPort       = 8000
	defaultPrometheusPort = 9000
)

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
	Config     ServerConfig
	Prometheus ServerConfig

	// Registery allows Lile to work with external registeries like
	// consul, zookeeper or similar
	Registery Registery

	// Private utils, exposed so they can be useful if needed
	ServiceListener  net.Listener
	GRPCServer       *grpc.Server
	PrometheusServer *http.Server
}

type RegisterImplementation func(s *grpc.Server)

// ServerConfig is a generic server configuration
type ServerConfig struct {
	Port int
	Host string
}

func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func defaultOptions(n string) *Service {
	return &Service{
		ID:                 generateID(n),
		Name:               n,
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

func generateID(n string) string {
	uid, _ := uuid.NewV4()
	return n + "-" + uid.String()
}

func BaseCommand(serviceName, shortDescription string) *cobra.Command {
	command := &cobra.Command{
		Use:   serviceName,
		Short: shortDescription,
	}

	command.PersistentFlags().StringVar(&grpcHost, "grpc_host", "", "gRPC service hostname")
	command.PersistentFlags().IntVar(&grpcPort, "grpc_port", 0, "gRPC port")
	command.PersistentFlags().StringVar(&prometheusHost, "prometheus_host", "", "Prometheus metrics hostname")
	command.PersistentFlags().IntVar(&prometheusPort, "prometheus_port", 0, "Prometheus metrics port")

	return command
}

// GlobalService returns the global service
func GlobalService() *Service {
	return service
}

// SetGlobalService returns the global service
func SetGlobalService(s *Service) *Service {
	service = s
	return s
}

// NewService creates a lile service with default options
func NewService(name string) *Service {
	return defaultOptions(name)
}

// Register attaches the gRPC implementation to the service
func (s *Service) Register(r func(s *grpc.Server)) {
	s.GRPCImplementation = r
}

// AddUnaryInterceptor adds a unary interceptor to the RPC server
func (s *Service) AddUnaryInterceptor(unint grpc.UnaryServerInterceptor) {
	s.UnaryInts = append(s.UnaryInts, unint)
}

// AddStreamInterceptor adds a stream interceptor to the RPC server
func (s *Service) AddStreamInterceptor(sint grpc.StreamServerInterceptor) {
	s.StreamInts = append(s.StreamInts, sint)
}

// URLForService returns a service URL via a registry or a simple DNS name
// if not available via the registery
func (s *Service) URLForService(name string) string {
	if s.Registery != nil {
		registeryURL, err := s.Registery.Get(name)
		if err != nil {
			fmt.Printf("lile: error contacting Registery for service %s. err: %s \n", name, err.Error())
		}

		return registeryURL
	}

	return name + ":80"
}

func (s *Service) setConfigFromFlags() {
	if grpcHost == "" {
		s.Config.Host = defaultHost
		s.Config.Port = defaultGRPCPort
	}

	if prometheusHost == "" {
		s.Prometheus.Host = defaultHost
		s.Prometheus.Port = defaultPrometheusPort
	}
}
