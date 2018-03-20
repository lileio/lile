package models

import (
	"fmt"

	"github.com/lileio/pubsub"
	"google.golang.org/grpc"
)

// Service is a grpc compatible server with extra features
type Service struct {
	Name string
	ID   string
	Subscriber pubsub.Subscriber
	// Interceptors
	UnaryInts  []grpc.UnaryServerInterceptor
	StreamInts []grpc.StreamServerInterceptor
	// The RPC server implementation
	GRPCImplementation RegisterImplementation
	Config             AppConfig
}

func (s *Service) Address() string {
	return s.Config.Service.Address()
}

type RegisterImplementation func(s *grpc.Server)

// A generic http server configuration
type HttpConfig struct {
	Port int
	Host string
}

func (c *HttpConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// The application config holder
type AppConfig struct {
	RegistryAddress string
	RegistryProvider string
	PubSubProvider string
	PubSubAddress string
	// Service host configuration
	Service         HttpConfig
	Prometheus      HttpConfig
}

// Checks whether the services uses a registry.
func (c *AppConfig) UsesRegistry() bool {
	return c.RegistryProvider != ""
}

// Checks whether the service uses pubsub.
func (c *AppConfig) UsesPubSub() bool {
	return c.PubSubProvider != ""
}