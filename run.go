package lile

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func (s *Service) Run() {
	s.setConfigFromFlags()

	if s.Registery != nil {
		s.Registery.Register(s)
	}

	errChan := make(chan error)

	go func(e chan<- error) {
		e <- s.ServeGRPC()
	}(errChan)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if s.Registery != nil {
			s.Registery.DeRegister(s)
		}

		logrus.Fatalf("Application error: %v", err)
	case sig := <-signalChan:
		logrus.Infof("Caught %v, attempting graceful shutdown...", sig)
		s.shutdown()
		os.Exit(0)
	}
}

func (s *Service) ServeGRPC() error {
	var err error
	s.ServiceListener, err = net.Listen("tcp", s.Config.Address())
	if err != nil {
		return err
	}

	logrus.Infof("Serving gRPC on %s", s.Config.Address())
	return s.createGrpcServer().Serve(s.ServiceListener)
}

func (s *Service) createGrpcServer() *grpc.Server {
	s.GRPCOptions = append(s.GRPCOptions, grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(service.UnaryInts...)))

	s.GRPCOptions = append(s.GRPCOptions, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(service.StreamInts...)))

	s.GRPCServer = grpc.NewServer(
		s.GRPCOptions...,
	)

	s.Register(s.GRPCImplementation)

	grpc_prometheus.EnableHandlingTimeHistogram(
		func(opt *prometheus.HistogramOpts) {
			opt.Buckets = prometheus.ExponentialBuckets(0.005, 1.4, 20)
		},
	)

	grpc_prometheus.Register(s.GRPCServer)

	s.startPrometheusServer()
	return s.GRPCServer
}

func (s *Service) startPrometheusServer() {
	s.PrometheusServer = &http.Server{Addr: s.Prometheus.Address()}

	http.Handle("/metrics", promhttp.Handler())
	logrus.Infof("Prometheus metrics at http://%s/metrics", s.Prometheus.Address())

	go func() {
		if err := s.PrometheusServer.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			logrus.Errorf("Prometheus http server: ListenAndServe() error: %s", err)
		}
	}()
}

func (s *Service) shutdown() {
	if s.Registery != nil {
		s.Registery.DeRegister(s)
	}

	s.GRPCServer.GracefulStop()

	// 30 seconds is the default grace period in Kubernetes
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	if err := s.PrometheusServer.Shutdown(ctx); err != nil {
		logrus.Infof("Timeout during shutdown of metrics server. Error: %v", err)
	}
}
