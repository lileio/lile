package lile

import (
	"net"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
)

// NewTestServer is a helper function to create a gRPC server on a non-network
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

// Creates a server listener dependent on the underlying platform. Windows
// hosts will have a Windows Named pipe, anything else gets a UNIX socket
func getTestServerTransport() (string, net.Listener, error) {
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

func generateID(n string) string {
	uid, _ := uuid.NewV4()
	return n + "-" + uid.String()
}
