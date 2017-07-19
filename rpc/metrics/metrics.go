//go:generate mockgen -source metrics.go -destination metric_mock.go -package metrics
package metrics

import (
	"google.golang.org/grpc"
)

// Monitor is an interface to metrics providers (i.e prometheus)
type Monitor interface {
	// Intercept allows a monitor to add gRPC interceptors before the
	// gRPC server is fully intialized
	InterceptRPC() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor)

	// Register takes a configured gRPC server and provides metrics.
	Register(*grpc.Server) error
}
