package lile

import (
	"net"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

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
