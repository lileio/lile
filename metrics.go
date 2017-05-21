//go:generate mockgen -source metrics.go -destination metric_mock.go -package lile
package lile

import (
	"google.golang.org/grpc"
)

// Monitor is an interface to metrics providers (i.e prometheus)
type Monitor interface {
	// Intercept allows a monitor to add gRPC interceptors before the
	// gRPC server is fully intialized
	InterceptRPC() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor)

	// Register takes a configured lile Service and provides metrics.
	// Serving or saving those metrics is usually done in a goroutine.
	// Returning an error will cause a fatal start for a lile service.
	Register(*Service) error
}
