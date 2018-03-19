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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/hashicorp/consul/api"
	"github.com/lileio/lile/fromenv"
	"github.com/lileio/pubsub"
	"github.com/lileio/lile/registry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	service = NewService("lile")
	serviceListener net.Listener
	prometheusServer *http.Server
	grpcServer *grpc.Server
)

type registerImplementation func(s *grpc.Server)

type HttpConfig struct {
	Port int
	Host string
}

func (c *HttpConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type AppConfig struct {
	RegistryAddress string
	UseConsul       bool
	UsePubSub       bool
	Service         HttpConfig
	Prometheus      HttpConfig
}

// Service is a grpc compatible server with extra features
type Service struct {
	Name string
	ID   string
	Subscriber pubsub.Subscriber
	// Interceptors
	UnaryInts  []grpc.UnaryServerInterceptor
	StreamInts []grpc.StreamServerInterceptor
	// The RPC server implementation
	GRPCImplementation registerImplementation
	Config             AppConfig
}

func (s *Service) Address() string {
	return s.Config.Service.Address()
}

func BaseCommand(serviceName, shortDescription string) *cobra.Command {
	command := &cobra.Command{
		Use:   serviceName,
		Short: shortDescription,
	}
	command.PersistentFlags().StringVar(&service.Config.RegistryAddress, "registry_address", "localhost:8500", "Address to use for consul.")
	command.PersistentFlags().BoolVar(&service.Config.UseConsul, "consul", false, "Enables consul to be used for registry.")
	command.PersistentFlags().BoolVar(&service.Config.UsePubSub, "pubsub", false, "Enables publish-subscribe.")
	command.PersistentFlags().IntVar(&service.Config.Prometheus.Port, "prometheus_port", 9000, "Prometheus port.")
	command.PersistentFlags().StringVar(&service.Config.Prometheus.Host, "prometheus_host", "localhost", "Prometheus hostname.")
	command.PersistentFlags().IntVar(&service.Config.Service.Port, "service_port", 8000, "Service port.")
	command.PersistentFlags().StringVar(&service.Config.Service.Host, "service_host", "localhost", "Service hostname.")

	return command
}

func generateId(n string) string {
	uid, _ := uuid.NewV4()
	return n + "-" + uid.String()
}

func defaultOptions(n string) Service {

	return Service{
		ID: generateId(n),
		Name: n,
		GRPCImplementation: func(s *grpc.Server) {},
	}
}

// NewService creates a lile service with N options
func NewService(name string) Service {
	return defaultOptions(name)
}

func Subscriber(s pubsub.Subscriber) {
	service.Subscriber = s
}

func GlobalService() *Service {
	return &service
}

func Id(n string) {
	service.ID = n
}

func Name(n string) {
	service.Name = n
	service.ID = generateId(n)
}

func Port(n int) {
	service.Config.Service.Port = n
}

func Host(h string) {
	service.Config.Service.Host = h
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

func ShutdownHook() {
	errChan := make(chan error, 10)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				logrus.Fatalf("error during application: %v", err)
			}
		case s := <-signalChan:
			logrus.Infof("Captured %v. Exiting...", s)
			shutdown()
			os.Exit(0)
		}
	}
}

func shutdown() {
	logrus.Infof("Shutting gRPC service '%s'", service.Name)
	if service.Config.UseConsul {
		rc, rcErr := getRegistryClient()
		if rcErr == nil {
			err := rc.DeRegister(service.Name)
			if err != nil {
				logrus.Errorf("Failed to deregister service by id: '%s'. Error: %v", service.ID, err)
			} else {
				logrus.Infof("Deregistered service '%s' at consul.", service.ID)
			}
		}
	}

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.TODO(), 5 *time.Second)
	defer cancel()
	if err := prometheusServer.Shutdown(ctx); err != nil {
		logrus.Fatalf("Timed out during shutdown of prometheus server. Error: %v", err)
	}

	if service.Config.UsePubSub {
		logrus.Infof("Unsubscribing...")
		// TODO: unsubscribe pubsub.Unsubscribe(service.Subscriber)
		// right now signal takes care of stopping the pubsub.Subscribe goroutine
	}
}

func Run() error {
	var err error
	serviceListener, err = net.Listen("tcp", service.Address())
	if err != nil {
		return err
	}

	// add to registry
	if service.Config.UseConsul {
		rc, rcErr := getRegistryClient()
		if rcErr == nil {
			err := rc.Register(service.ID, service.Name, service.Config.Service.Port, &api.AgentServiceCheck{
				CheckID:                        "",
				Name:                           "",
				Args:                           nil,
				Script:                         "",
				DockerContainerID:              "",
				Shell:                          "",
				Interval:                       "",
				Timeout:                        "",
				TTL:                            "",
				HTTP:                           "",
				Header:                         nil,
				Method:                         "",
				TCP:                            "",
				Status:                         "",
				Notes:                          "",
				TLSSkipVerify:                  false,
				GRPC:                           "",
				GRPCUseTLS:                     false,
				DeregisterCriticalServiceAfter: "",
			})

			if err != nil {
				logrus.Errorf("Failed to register service at '%s'. error: %v", service.Config.RegistryAddress, err)
			} else {
				logrus.Infof("Regsitered service '%s' at consul.", service.ID)
			}
		} else {
			logrus.Errorf("Failed to create registry client with address: '%s'. Error: %v", service.Config.RegistryAddress, err)
		}
	}

	// setup pubsub
	if service.Config.UsePubSub {
		logrus.Infof("Subscribing...")
		// TODO: make Subscribe return without waiting for signal
		go pubsub.Subscribe(service.Subscriber)
	}

	logrus.Infof("Serving gRPC on %s", service.Address())
	return CreateServer().Serve(serviceListener)
}

func getRegistryClient() (registry.Client, error) {
	rc, err := registry.NewConsulClient(service.Config.RegistryAddress)
	if err != nil {
		logrus.Errorf("Failed to create wuth registry address '%s'", service.Config.RegistryAddress)
		return nil, err
	}

	return rc, nil
}

func CreateServer() *grpc.Server {
	zap, err := zap.NewProduction()
	if err == nil {
		AddStreamInterceptor(grpc_zap.StreamServerInterceptor(zap))
	}
	// Default interceptors, [prometheus, opentracing]
	AddUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor)
	AddStreamInterceptor(grpc_prometheus.StreamServerInterceptor)
	AddUnaryInterceptor(otgrpc.OpenTracingServerInterceptor(
		fromenv.Tracer(service.Name)))

	// add recovery later to avoid panics within handlers
	AddStreamInterceptor(grpc_recovery.StreamServerInterceptor())
	AddUnaryInterceptor(grpc_recovery.UnaryServerInterceptor())

	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(service.UnaryInts...)),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(service.StreamInts...)),
	)

	service.GRPCImplementation(grpcServer)

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(grpcServer)

	startPrometheusServer()

	return grpcServer
}

func startPrometheusServer() {

	prometheusServer = &http.Server{Addr: service.Config.Prometheus.Address() }

	http.Handle("/metrics", promhttp.Handler())

	logrus.Infof("Prometheus metrics at http://%s/metrics", service.Config.Prometheus.Address())

	go func() {
		if err := prometheusServer.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			logrus.Errorf("Prometheus http server: ListenAndServe() error: %s", err)
		}
	}()
}

// NewTestServer is a helper function to create a gRPC server on a unix socket
// it returns the socket location and a func to call which starts the server
func NewTestServer(s *grpc.Server) (string, func()) {
	// Create a temp random unix socket
	uid, err := uuid.NewV1()
	if err != nil {
		panic(err)
	}

	skt := "/tmp/" + uid.String()

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
