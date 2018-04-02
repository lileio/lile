// Package lile provides helper methods to quickly create RPC based services
// that have metrics, tracing and pub/sub support
package lile

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/lile/fromenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	service = NewService("lile")
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

func BaseCommand(serviceName, shortDescription string) *cobra.Command {
	command := &cobra.Command{
		Use:   serviceName,
		Short: shortDescription,
	}

	command.PersistentFlags().StringVar(&service.Config.Host, "grpc_host", "0.0.0.0", "gRPC service hostname")
	command.PersistentFlags().IntVar(&service.Config.Port, "grpc_port", 8000, "gRPC port")
	command.PersistentFlags().StringVar(&service.Prometheus.Host, "prometheus_host", "0.0.0.0", "Prometheus metrics hostname")
	command.PersistentFlags().IntVar(&service.Prometheus.Port, "prometheus_port", 9000, "Prometheus metrics port")

	return command
}

// GlobalService returns the global service
func GlobalService() *Service {
	return &service
}

// NewService creates a lile service with default options
func NewService(name string) Service {
	return defaultOptions(name)
}

// Register attaches the gRPC implementation to the service
func (s *Service) Register(r RegisterImplementation) {
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

func (s *Service) Run() {
	if s.Registery != nil {
		s.Registery.Register(s)
	}

	errChan := make(chan error)

	go func(e chan<- error) {
		e <- s.ServeGRPC()
	}(errChan)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if s.Registery != nil {
			s.Registery.DeRegister(s)
		}

		logrus.Fatalf("Application startup error: %v", err)
	case sig := <-signalChan:
		logrus.Infof("Caught %v, attempting graceful shutdown...", sig)
		s.shutdown()
		os.Exit(0)
	}
}

func (s *Service) shutdown() {
	if s.Registery != nil {
		s.Registery.DeRegister(s)
	}

	s.GRPCServer.GracefulStop()

	// 30 seconds is the default grace period in Kubernetes
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	if err := s.PrometheusServer.Shutdown(ctx); err != nil {
		logrus.Infof("Timeout during shutdown of metrics server. Error: %v", err)
	}
}

func (s *Service) ServeGRPC() error {
	var err error
	s.ServiceListener, err = net.Listen("tcp", s.Config.Address())
	if err != nil {
		return err
	}

	logrus.Infof("Serving gRPC on %s", s.Config.Address())
	return s.createGrpcServer().Serve(s.ServiceListener)
}

func (s *Service) createGrpcServer() *grpc.Server {
	s.GRPCOptions = append(s.GRPCOptions, grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(service.UnaryInts...)))

	s.GRPCOptions = append(s.GRPCOptions, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(service.StreamInts...)))

	s.GRPCServer = grpc.NewServer(
		s.GRPCOptions...,
	)

	s.Register(s.GRPCImplementation)

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(s.GRPCServer)

	s.startPrometheusServer()
	return s.GRPCServer
}

func (s *Service) startPrometheusServer() {
	s.PrometheusServer = &http.Server{Addr: s.Prometheus.Address()}

	http.Handle("/metrics", promhttp.Handler())
	logrus.Infof("Prometheus metrics at http://%s/metrics", s.Prometheus.Address())

	go func() {
		if err := s.PrometheusServer.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			logrus.Errorf("Prometheus http server: ListenAndServe() error: %s", err)
		}
	}()
}

func defaultOptions(n string) Service {
	return Service{
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
