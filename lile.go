// Package lile provides helper methods to quickly create RPC based services
// that have metrics, tracing and pub/sub support
package lile

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/lile/pubsub"
	"github.com/lileio/lile/rpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

// Service is a grpc compatible server with extra features
type Service struct {
	Name string

	// Opentracing based tracer
	Tracer opentracing.Tracer

	// PubSub for publish and subscribe async messaging
	PubSubProvider   pubsub.Provider
	PubSubClient     pubsub.Client
	PubSubSubscriber pubsub.Subscriber

	// RPC
	RPCOptions rpc.RPCOptions
	RPCServer  *grpc.Server
}

// NewService creates a lile service with N options
func NewService(name string, opts ...interface{}) *Service {
	// Setup a Service with default configuration
	s := &Service{
		Name:       name,
		Tracer:     opentracing.GlobalTracer(),
		RPCOptions: rpc.DefaultRPCOptions(),
	}

	// Loop through options and apply setup for each type
	for _, opt := range opts {
		switch o := opt.(type) {
		case rpc.RPCOption:
			o(&s.RPCOptions)
		case pubsub.Subscriber:
			s.PubSubSubscriber = o
		case pubsub.Provider:
			s.PubSubProvider = o
		case opentracing.Tracer:
			s.Tracer = o
		default:
			fmt.Printf("o = %+v\n", o)
			log.Fatalf("lile: NewService cannot accept option %T", opt)
		}
	}

	// Setup tracing
	rpc.AddUnaryInterceptor(
		otgrpc.OpenTracingServerInterceptor(s.Tracer),
	)(&s.RPCOptions)

	// Setup Publish Subscribe
	s.PubSubClient = pubsub.Client{Provider: s.PubSubProvider}

	// Automatic publishers
	if len(s.RPCOptions.PublishMethods) != 0 {
		if s.PubSubProvider == nil {
			logrus.Fatalf("lile pubsub: publishers specified but no Provider is set/available")
		}

		rpc.AddUnaryInterceptor(
			pubsub.UnaryServerInterceptor(&s.PubSubClient, s.RPCOptions.PublishMethods),
		)(&s.RPCOptions)
	}

	// Setup server
	s.RPCServer = rpc.NewRPCServer(s.RPCOptions)

	return s
}

func (s Service) Subscribe() {
	logrus.Info("lile pubsub: Subscribed to events")
	s.PubSubSubscriber.Setup(&s.PubSubClient)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

func (s Service) ServeRPCInsecure() error {
	return rpc.ListenAndServeInsecure(s.RPCServer, s.RPCOptions)
}

func (s Service) RPCTestServer() (string, func()) {
	return rpc.NewTestServer(s.RPCServer)
}
