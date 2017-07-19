package prometheus

import (
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

func NewMonitor() Prometheus {
	return Prometheus{}
}

type Prometheus struct{}

func (p Prometheus) InterceptRPC() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	return grpc_prometheus.UnaryServerInterceptor,
		grpc_prometheus.StreamServerInterceptor
}

func (p Prometheus) Register(s *grpc.Server) error {
	grpc_prometheus.Register(s)
	return nil
}
