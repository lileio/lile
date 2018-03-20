package registry

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

const (
	CONSUL_DEFAULT_ADDRESS    = "localhost:8500"
	ZOOKEEPER_DEFAULT_ADDRESS = "localhost:3000"
)

type ServiceDescriptor struct {
}

type MetaData struct {
}

type Check struct {
}

//Client provides an interface for getting data out of Consul
type Client interface {
	// Get a Service from consul
	Service(string, string, *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)
	// Register a service with local agent
	Register(id string, serviceName string, port int, check *api.AgentServiceCheck) error
	// Deregister a service with local agent
	DeRegister(string) error
}

func NewRegistryClient(provider, address string) (Client, error) {
	switch provider {
	case "consul":
		addr := address
		if len(addr) == 0 {
			addr = CONSUL_DEFAULT_ADDRESS
		}
		return NewConsulClient(addr)
	case "zookeeper":
		return nil, errors.New("zookeeper not implemented yet")
	default:
		return nil, fmt.Errorf("unknown registry provider: %s", provider)
	}
}
