package zookeeper

import "github.com/pkg/errors"

// TODO: zookeeper

const ZOOKEEPER_DEFAULT_ADDRESS = "localhost:3000"

type registryClient struct {
	//consul *zookeeper.Client
	address string
}

func NewClient(addr string) (*registryClient, error) {
	return nil, errors.New("not implemented yet")
}
