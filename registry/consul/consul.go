package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

const CONSUL_DEFAULT_ADDRESS = "localhost:8500"

type registryClient struct {
	consul  *api.Client
	address string
}

// NewClient returns a Client interface for given consul address
func NewClient(addr string) (*registryClient, error) {
	if len(addr) == 0 {
		addr = CONSUL_DEFAULT_ADDRESS
	}
	config := api.DefaultConfig()
	config.Address = addr
	c, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &registryClient{consul: c, address: addr}, nil
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
		Tags:    nil,
		Address: "",
	}
	err := c.consul.Agent().ServiceRegister(reg)
	if err != nil {
		logrus.Errorf("Failed to register service at '%s'. error: %v", c.address, err)
	} else {
		logrus.Infof("Regsitered service '%s' at consul.", id)
	}
	return err
}

// DeRegister a service with consul local agent
func (c *registryClient) DeRegister(id string) error {
	err := c.consul.Agent().ServiceDeregister(id)
	if err != nil {
		logrus.Errorf("Failed to deregister service by id: '%s'. Error: %v", id, err)
	} else {
		logrus.Infof("Deregistered service '%s' at consul.", id)
	}
	return err
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
