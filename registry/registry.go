package registry

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/lileio/lile/consul"
	"github.com/pkg/errors"
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
		return consul.NewClient(address)
	case "zookeeper":
		return nil, errors.New("not implemented yet")
	default:
		return nil, fmt.Errorf("unknown registry provider: %s", provider)
	}
}
