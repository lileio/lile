package prometheus

import (
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/lileio/lile"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

func NewMonitor() Prometheus {
	return Prometheus{}
}

type Prometheus struct {
	Port string
	Addr string
}

func (p Prometheus) InterceptRPC() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	return grpc_prometheus.UnaryServerInterceptor,
		grpc_prometheus.StreamServerInterceptor
}

func (p Prometheus) Register(s *lile.Service) error {
	if p.Port == "" {
		p.Port = ":9000"
	}

	if p.Addr == "" {
		p.Addr = "/"
	}

	grpc_prometheus.Register(s.RPCServer)

	mux := http.NewServeMux()
	mux.Handle(p.Addr, prometheus.Handler())
	go http.ListenAndServe(p.Port, mux)
	return nil
}
