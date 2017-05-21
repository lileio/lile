package rpc

import (
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type registerImplementation func(s *grpc.Server)

type RPCOption func(*RPCOptions)

type RPCOptions struct {
	Port string

	// Interceptors
	UnaryInts  []grpc.UnaryServerInterceptor
	StreamInts []grpc.StreamServerInterceptor

	// The RPC server implementation
	GRPCImplementation registerImplementation

	// Methods that automatically publish to lile's pubsub
	// format being map[methodName]topic...
	// map{"UpdateBooking": "bookings.updated"}
	PublishMethods map[string]string
}

func DefaultRPCOptions() RPCOptions {
	return RPCOptions{
		Port:               ":8000",
		GRPCImplementation: func(s *grpc.Server) {},
		PublishMethods:     map[string]string{},
	}
}

// RPCPort sets the gRPC port of the service
func RPCPort(n string) RPCOption {
	return func(o *RPCOptions) {
		o.Port = n
	}
}

// AddUnaryInterceptor adds a unary interceptor to the RPC server
func AddUnaryInterceptor(unint grpc.UnaryServerInterceptor) RPCOption {
	return func(o *RPCOptions) {
		o.UnaryInts = append(o.UnaryInts, unint)
	}
}

// AddStreamInterceptor adds a stream interceptor to the RPC server
func AddStreamInterceptor(sint grpc.StreamServerInterceptor) RPCOption {
	return func(o *RPCOptions) {
		o.StreamInts = append(o.StreamInts, sint)
	}
}

// Implementation registers the server handler for RPC calls
func Implementation(impl registerImplementation) RPCOption {
	return func(o *RPCOptions) {
		o.GRPCImplementation = impl
	}
}

// AddPublishMethod registers a method for automatic pubsub publishing via
// an interceptor, currently only unary methods are supported
func AddPublishMethod(method, topic string) RPCOption {
	return func(o *RPCOptions) {
		o.PublishMethods[method] = topic
	}
}

func NewRPCServer(opts RPCOptions) *grpc.Server {
	gs := grpc.NewServer(
		// Interceptors
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(opts.UnaryInts...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(opts.StreamInts...)),
	)

	opts.GRPCImplementation(gs)

	return gs
}

// ListenAndServeInsecure creates a tcp socket and starts listening for connections.
func ListenAndServeInsecure(s *grpc.Server, o RPCOptions) error {
	lis, err := net.Listen("tcp", o.Port)
	if err != nil {
		return err
	}

	logrus.Infof("Serving gRPC on %s", o.Port)
	return s.Serve(lis)
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
