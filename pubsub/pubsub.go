//go:generate mockgen -source pubsub.go -destination pubsub_mock.go -package pubsub
package pubsub

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
)

var client = &Client{Provider: NoopProvider{}}

var (
	publishedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pubsub_message_published_total",
			Help: "Total number of messages published by the client.",
		}, []string{"topic", "service"})

	publishedSize = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pubsub_outgoing_bytes",
			Help: "A counter of pubsub published outgoing bytes.",
		}, []string{"topic", "service"})
)

func init() {
	prometheus.MustRegister(publishedCounter)
	prometheus.MustRegister(publishedSize)
}

// Client holds a reference to a Provider
type Client struct {
	ServiceName string
	Provider    Provider
}

func SetClient(c *Client) {
	client = c
}

func GlobalClient() *Client {
	return client
}

// Provider is generic interface for a pub sub provider
type Provider interface {
	Publish(ctx context.Context, topic string, msg proto.Message) error
	Subscribe(topic, subscriberName string, h MsgHandler, deadline time.Duration, autoAck bool)
}

func Publish(ctx context.Context, topic string, msg proto.Message) error {
	err := client.Provider.Publish(ctx, topic, msg)
	if err != nil {
		return err
	}

	publishedCounter.WithLabelValues(topic, client.ServiceName).Inc()
	publishedSize.WithLabelValues(topic, client.ServiceName).Add(float64(len([]byte(msg.String()))))
	return nil
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

type NoopProvider struct{}

func (np NoopProvider) Publish(ctx context.Context, topic string, msg proto.Message) error {
	return nil
}

func (np NoopProvider) Subscribe(topic, subscriberName string, h MsgHandler, deadline time.Duration, autoAck bool) {
	return
}
