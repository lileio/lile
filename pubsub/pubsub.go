package pubsub

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
)

// Client holds a reference to a Provider
type Client struct {
	Provider Provider

	// A map of gRPC method names to automatically publish for and their topic
	InterceptorMethods map[string]string
}

// Provider is generic interface for a pub sub provider
type Provider interface {
	Publish(ctx context.Context, topic string, msg proto.Message) error
	Subscribe(topic string, h MsgHandler, deadline time.Duration, autoAck bool)
}

// Subscriber is a service/service that listens to events and registers handlers
// for those events
type Subscriber interface {
	// Setup is a required method that allows the subscriber service to add handlers
	// and perform any setup if required, this is usually called by lile upon start
	Setup(*Client)
}

// Msg is a lile representation of a pub sub message
type Msg struct {
	ID       string
	Metadata map[string]string
	Data     []byte

	Ack  func()
	Nack func()
}

// Handler is a specific callback used for Subscribe. It is generalized to
// an interface{}, but we will discover its format and arguments at runtime
// and perform the correct callback, including de-marshalling of protobuf.
//
// Handlers are expected to have one of four signatures.
//
//	type person struct {
//		Name string
//		Age  uint
//	}
//
//	handler := func(ctx context.Context, m *pubsub.Msg)
//	handler := func(p *person)
//	handler := func(ctx context.Context, p *person)
//	handler := func(ctx context.Context, metadata map[string]string, p *person)
//
// These forms allow a callback to request a raw Msg where the processing
// is untouched or to have the library perform some de-marshalling.
// It is highly recommended to use context when you call external services so
// that tracing can be kept even for async actions
type Handler interface{}

// MsgHandler is the internal or raw message handler
type MsgHandler func(c context.Context, m Msg) error
