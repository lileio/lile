package registry

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type registryClient struct {
	consul *api.Client
}

//NewConsul returns a Client interface for given consul address
func NewConsulClient(addr string) (*registryClient, error) {
	config := api.DefaultConfig()
	config.Address = addr
	c, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &registryClient{consul: c}, nil
}

// Register a service with consul local agent
func (c *registryClient) Register(id string, name string, port int, check *api.AgentServiceCheck) error {
	reg := &api.AgentServiceRegistration{
		ID:                name,
		Name:              name,
		EnableTagOverride: false,
		Check:             check,
		Port:              port,
		// TODO: fill tags and address (os.getenv(hostname))
		Tags:              nil,
		Address:           "",
	}
	return c.consul.Agent().ServiceRegister(reg)
}

// DeRegister a service with consul local agent
func (c *registryClient) DeRegister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// Service return a service
func (c *registryClient) Service(service, tag string, queryOpts *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.consul.Health().Service(service, tag, passingOnly, queryOpts)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}
