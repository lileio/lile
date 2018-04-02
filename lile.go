// Package lile provides helper methods to quickly create RPC based services
// that have metrics, tracing and pub/sub support
package lile

import (
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/lile/fromenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/natefinch/npipe"
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

//   Creates a server listener dependent on the underlying platform. Windows
// hosts will have a Windows Named pipe, anything else gets a UNIX socket
func getTestServerTransportOrPanic()(string, net.Listener) {
	var uniqueAddress string
	var serverListener net.Listener
	var err error

	// Create a temp random unix socket
	uid, err := uuid.NewV1()
	if err != nil {
		panic(err)
	}
	uniqueAddress = uid.String()

	if runtime.GOOS == "windows" {
		uniqueAddress = `\\.\pipe\` + uniqueAddress
		serverListener, err = npipe.Listen(uniqueAddress)

	} else {
		uniqueAddress = "/tmp/" + uniqueAddress
		serverListener, err = net.Listen("unix", uniqueAddress)

	}

	if err != nil {
		panic(err)
	}

	return uniqueAddress, serverListener
}

//   NewTestServer is a helper function to create a gRPC server on a non-network
// socket and it returns the socket location and a func to call which starts
// the server
func NewTestServer(s *grpc.Server) (string, func()) {

	socketAddress, listener := getTestServerTransportOrPanic()

	return socketAddress, func() {
		s.Serve(listener)
	}
}

//   Returns a dialer function for the underlying platform. Returns a Windows
// named pipe if asked for Windows, else a UNIX socket
func getDialerFunctionForPlatform(platform string)(
		func(string)(net.Conn, error)) {

	var dialFunc func(string)(net.Conn, error)

	if platform == "windows" {
		dialFunc = func(addr string)(net.Conn, error) {
			return npipe.Dial(addr)
		}
	} else {
		dialFunc = func(addr string)(net.Conn, error) {
			return net.Dial("unix", addr)
		}
	}

	return dialFunc
}

// TestConn is a connection that connects to a socket based connection
func TestConn(addr string) *grpc.ClientConn {

	dialFunc := getDialerFunctionForPlatform(runtime.GOOS)

	conn, err := grpc.Dial(
		addr,
		grpc.WithDialer(func(addr string, d time.Duration) (net.Conn, error) {
			return dialFunc(addr)
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
