package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
)

// Memory is an in memory pubsub system for testing only
type Memory struct {
	Messages map[string][]Msg
}

func NewMemory() *Memory {
	return &Memory{
		Messages: map[string][]Msg{},
	}
}

func (m *Memory) Publish(ctx context.Context, topic string, msg proto.Message) error {
	fmt.Printf("m = %+v\n", m)
	return nil
}

func (m *Memory) Subscribe(topic string, h MsgHandler, deadline time.Duration, autoAck bool) {
}
