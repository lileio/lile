package lile_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lileio/lile"
	"github.com/lileio/lile/pubsub"
	"github.com/lileio/lile/rpc"
	"github.com/stretchr/testify/assert"
)

func TestBlankNewService(t *testing.T) {
	s := lile.NewService("test_service")
	assert.NotNil(t, s)
	assert.Equal(t, s.Name, "test_service")
}

func TestRPCOptions(t *testing.T) {
	s := lile.NewService("test_service", rpc.RPCPort(":9000"))
	assert.NotNil(t, s)
	assert.Equal(t, s.RPCOptions.Port, ":9000")
}

func TestMonitor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMonitor := lile.NewMockMonitor(ctrl)
	mockMonitor.EXPECT().InterceptRPC()
	mockMonitor.EXPECT().Register(gomock.Any())

	s := lile.NewService("test_service", mockMonitor)
	assert.NotNil(t, s)
	assert.NotNil(t, s.Monitor)
}

func TestPubSub(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mp := pubsub.NewMockProvider(ctrl)

	s := lile.NewService("test_service", mp)
	assert.NotNil(t, s)
	assert.NotNil(t, s.PubSubClient)
}
