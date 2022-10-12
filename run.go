package lile

import (
	"context"
	"github.com/lileio/pubsub/v2"
	"go.uber.org/fx"
	"net"
	"net/http"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Run is a blocking cmd to run the gRPC and metrics server.
// You should listen to os signals and call Shutdown() if you
// want a graceful shutdown or want to handle other goroutines
func Run(lifecycle fx.Lifecycle) {

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if service.Registry != nil {
				service.Registry.Register(service)
			}

			// Start a metrics server in the background
			startPrometheusServer()

			// Create and then server a gRPC server
			err := ServeGRPC()
			if service.Registry != nil {
				service.Registry.DeRegister(service)
			}
			return err
		},
		OnStop: func(ctx context.Context) error {
			if err := Shutdown(ctx); err != nil {
				return err
			}

			pubsub.Shutdown()
			return nil
		},
	})

}

// ServeGRPC creates and runs a blocking gRPC server
func ServeGRPC() error {
	var err error
	service.ServiceListener, err = net.Listen("tcp", service.Config.Address())
	if err != nil {
		return err
	}

	logrus.Infof("Serving gRPC on %s", service.Config.Address())
	return createGrpcServer().Serve(service.ServiceListener)
}

// Shutdown gracefully shuts down the gRPC and metrics servers
func Shutdown(ctx context.Context) error {
	logrus.Infof("lile: Gracefully shutting down gRPC and Prometheus")

	if service.Registry != nil {
		if err := service.Registry.DeRegister(service); err != nil {
			return err
		}
	}

	service.GRPCServer.GracefulStop()

	// 30 seconds is the default grace period in Kubernetes
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)

	defer cancel()

	if err := service.PrometheusServer.Shutdown(ctx); err != nil {
		logrus.Infof("Timeout during shutdown of metrics server. Error: %v", err)
		return err
	}

	return nil
}

func createGrpcServer() *grpc.Server {
	service.GRPCOptions = append(service.GRPCOptions, grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(service.UnaryInts...)))

	service.GRPCOptions = append(service.GRPCOptions, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(service.StreamInts...)))

	service.GRPCServer = grpc.NewServer(
		service.GRPCOptions...,
	)

	service.GRPCImplementation(service.GRPCServer)

	grpc_prometheus.EnableHandlingTimeHistogram(
		func(opt *prometheus.HistogramOpts) {
			opt.Buckets = prometheus.ExponentialBuckets(0.005, 1.4, 20)
		},
	)

	grpc_prometheus.Register(service.GRPCServer)
	return service.GRPCServer
}

func startPrometheusServer() {
	service.PrometheusServer = &http.Server{Addr: service.PrometheusConfig.Address()}

	http.Handle("/metrics", promhttp.Handler())
	logrus.Infof("Prometheus metrics at http://%s/metrics", service.PrometheusConfig.Address())

	go func() {
		if err := service.PrometheusServer.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			logrus.Errorf("Prometheus http server: ListenAndServe() error: %s", err)
		}
	}()
}
