// Package lile provides helper methods to quickly create RPC based services
// that have metrics, tracing and pub/sub support
package lile

import (
	"context"
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
	"github.com/lileio/lile/registry"
	"github.com/lileio/pubsub"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	service          = NewService("lile")
	serviceListener  net.Listener
	prometheusServer *http.Server
	grpcServer       *grpc.Server
)

func BaseCommand(serviceName, shortDescription string) *cobra.Command {
	command := &cobra.Command{
		Use:   serviceName,
		Short: shortDescription,
	}
	command.PersistentFlags().StringVar(&service.Config.RegistryAddress, "registry_address", "", "Address to use for consul.")
	command.PersistentFlags().StringVar(&service.Config.RegistryProvider, "registry", "", "Sets the registry provider. Possible values: ['consul', 'zookeeper']")
	command.PersistentFlags().StringVar(&service.Config.PubSubProvider, "pubsub", "", "Sets the pubsub provider. Possible values: ['gcloud']")
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
		ID:                 generateId(n),
		Name:               n,
		GRPCImplementation: func(s *grpc.Server) {},
	}
}

// Returns the global service
func GlobalService() *Service {
	return &service
}

// NewService creates a lile service with N options
func NewService(name string) Service {
	return defaultOptions(name)
}

func Subscriber(s pubsub.Subscriber) {
	service.Subscriber = s
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

// Attaches the gRPC implementation to the service
func Server(r RegisterImplementation) {
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

func Run() {
	// add to registry
	if service.Config.UsesRegistry() {
		rc, rcErr := CreateRegistryClient()
		if rcErr == nil {
			rc.Register(service.ID, service.Name, service.Config.Service.Port, nil)
		} else {
			logrus.Errorf("Failed to create registry client with address: '%s'. Error: %v", service.Config.RegistryAddress, rcErr)
		}
	}

	// setup pubsub
	if service.Config.UsesPubSub() {
		logrus.Infof("Subscribing...")
		// TODO: make Subscribe return without waiting for signal
		go pubsub.Subscribe(service.Subscriber)
	}

	go func() {
		err := serveGrpc()
		if err != nil {
			logrus.Fatalf("gRPC serve failed. Error: %v", err)
		}
	}()

	gracefulShutdown()
}

func gracefulShutdown() {
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
	if service.Config.UsesRegistry() {
		rc, rcErr := CreateRegistryClient()
		if rcErr == nil {
			rc.DeRegister(service.Name)
		}
	}

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := prometheusServer.Shutdown(ctx); err != nil {
		logrus.Fatalf("Timeout during shutdown of prometheus server. Error: %v", err)
	}

	if service.Config.UsesPubSub() {
		logrus.Infof("Unsubscribing...")
		// TODO: unsubscribe pubsub.Unsubscribe(service.Subscriber)
		// right now signal takes care of stopping the pubsub.Subscribe goroutine
	}

	logrus.Infof("Application stopped.")
}

// Returns a registry client based on the config
func CreateRegistryClient() (registry.Client, error) {
	rc, err := registry.NewRegistryClient(service.Config.RegistryProvider, service.Config.RegistryAddress)
	if err != nil {
		logrus.Errorf("Failed to create wuth registry address '%s'", service.Config.RegistryAddress)
		return nil, err
	}

	return rc, nil
}

func serveGrpc() error {
	var err error
	serviceListener, err = net.Listen("tcp", service.Address())
	if err != nil {
		return err
	}

	logrus.Infof("Serving gRPC on %s", service.Address())

	return createGrpcServer().Serve(serviceListener)
}

func createGrpcServer() *grpc.Server {
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

	prometheusServer = &http.Server{Addr: service.Config.Prometheus.Address()}

	http.Handle("/metrics", promhttp.Handler())

	logrus.Infof("Prometheus metrics at http://%s/metrics", service.Config.Prometheus.Address())

	go func() {
		if err := prometheusServer.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			logrus.Errorf("Prometheus http server: ListenAndServe() error: %s", err)
		}
	}()
}

//   Creates a server listener dependent on the underlying platform. Windows
// hosts will have a Windows Named pipe, anything else gets a UNIX socket
func getTestServerTransport()(string, net.Listener, error) {
	var uniqueAddress string

	// Create a random string for part of the address
	uid, err := uuid.NewV1()
	if err != nil {
		return "", nil, err
	}

	uniqueAddress = formatPlatformTestSeverAddress(uid.String())

	serverListener, err := getTestServerListener(uniqueAddress)
	if err != nil {
		return "", nil, err
	}

	return uniqueAddress, serverListener, nil
}

//   NewTestServer is a helper function to create a gRPC server on a non-network
// socket and it returns the socket location and a func to call which starts
// the server
func NewTestServer(s *grpc.Server) (string, func()) {
	socketAddress, listener, err := getTestServerTransport()
	if err != nil {
		panic(err)
	}

	return socketAddress, func() {
		s.Serve(listener)
	}
}

// TestConn is a connection that connects to a socket based connection
func TestConn(addr string) *grpc.ClientConn {

	conn, err := grpc.Dial(
		addr,
		grpc.WithDialer(func(addr string, d time.Duration) (net.Conn, error) {
			return dialTestServer(addr)
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
