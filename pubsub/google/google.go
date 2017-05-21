package google

import (
	"context"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	ps "github.com/lileio/lile/pubsub"
	ctxNet "golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/jpillora/backoff"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
)

var mutex = &sync.Mutex{}
var pubsubTag = opentracing.Tag{string(ext.Component), "pubsub"}

type GoogleCloud struct {
	subName string
	client  *pubsub.Client
	topics  map[string]*pubsub.Topic
}

func NewGoogleCloud(project_id string, subName string) (*GoogleCloud, error) {
	ctx := ctxNet.Background()
	c, err := pubsub.NewClient(ctx, project_id)
	if err != nil {
		return nil, err
	}

	return &GoogleCloud{
		subName: subName,
		client:  c,
		topics:  map[string]*pubsub.Topic{},
	}, nil
}

func (g *GoogleCloud) Publish(ctx context.Context, topic string, msg proto.Message) error {
	var parentCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentCtx = parent.Context()
	}

	tracer := opentracing.GlobalTracer()
	clientSpan := tracer.StartSpan(
		topic,
		opentracing.ChildOf(parentCtx),
		ext.SpanKindProducer,
		pubsubTag,
	)

	defer clientSpan.Finish()

	b, err := proto.Marshal(msg)
	if err != nil {
		logrus.Errorf("Cant marshal msg for topic %s, err: %v", topic, err)
	}

	clientSpan.LogEvent("get topic")
	mutex.Lock()
	t, err := g.getTopic(topic)
	mutex.Unlock()
	clientSpan.LogEvent("topic received")
	if err != nil {
		return err
	}

	attrs := map[string]string{}
	tracer.Inject(
		clientSpan.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(attrs))

	clientSpan.LogEvent("publish")
	res := t.Publish(ctxNet.Background(), &pubsub.Message{
		Data:       b,
		Attributes: attrs,
	})

	_, err = res.Get(ctxNet.Background())
	clientSpan.LogEvent("publish confirmed")
	return err
}

func (g *GoogleCloud) Subscribe(topic string, h ps.MsgHandler, deadline time.Duration, autoAck bool) {
	g.subscribe(topic, h, deadline, autoAck, make(chan bool, 1))
}

func (g *GoogleCloud) subscribe(topic string, h ps.MsgHandler, deadline time.Duration, autoAck bool, ready chan<- bool) {
	go func() {
		var sub *pubsub.Subscription
		var err error
		b := &backoff.Backoff{
			Min: 500 * time.Millisecond,
			Max: 10 * time.Second,
		}

		// Subscribe with backoff for failure (i.e topic doesn't exist yet)
		for {
			t := g.client.Topic(topic)
			subName := g.subName + "--" + topic
			sc := pubsub.SubscriptionConfig{
				Topic:       t,
				AckDeadline: deadline,
			}
			sub, err = g.client.CreateSubscription(ctxNet.Background(), subName, sc)
			if err != nil && !strings.Contains(err.Error(), "AlreadyExists") {
				d := b.Duration()
				logrus.Errorf("Can't subscribe to topic: %s. Subscribing again in %s", err.Error(), d)
				time.Sleep(d)
				continue
			}

			b.Reset()
			logrus.Infof("Subscribed to topic %s with name %s", topic, subName)
			break
		}

		ready <- true

		// Listen to messages and call the MsgHandler
		for {
			err = sub.Receive(ctxNet.Background(), func(ctx ctxNet.Context, m *pubsub.Message) {
				logrus.Infof("Recevied on topic %s, id: %s", topic, m.ID)

				tracer := opentracing.GlobalTracer()
				spanContext, err := tracer.Extract(
					opentracing.TextMap,
					opentracing.TextMapCarrier(m.Attributes))
				if err != nil {
					logrus.Error(err)
					return
				}

				handlerSpan := tracer.StartSpan(
					g.subName,
					consumerOption{clientContext: spanContext},
					pubsubTag,
				)
				defer handlerSpan.Finish()
				ctx = opentracing.ContextWithSpan(ctx, handlerSpan)

				msg := ps.Msg{
					ID:       m.ID,
					Metadata: m.Attributes,
					Data:     m.Data,
					Ack: func() {
						m.Ack()
					},
					Nack: func() {
						m.Nack()
					},
				}

				err = h(ctx, msg)
				if err != nil {
					logrus.Error(err)
					return
				}

				if autoAck {
					m.Ack()
				}
			})

			if err != nil {
				logrus.Error(err)
			}
		}
	}()
}

type consumerOption struct {
	clientContext opentracing.SpanContext
}

func (c consumerOption) Apply(o *opentracing.StartSpanOptions) {
	if c.clientContext != nil {
		opentracing.ChildOf(c.clientContext).Apply(o)
	}
	ext.SpanKindConsumer.Apply(o)
}

func (g *GoogleCloud) getTopic(name string) (*pubsub.Topic, error) {
	if g.topics[name] != nil {
		return g.topics[name], nil
	}

	ctx := ctxNet.Background()
	t, err := g.client.CreateTopic(ctx, name)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		return nil, err
	}

	g.topics[name] = t

	return t, nil
}

func (g *GoogleCloud) deleteTopic(name string) error {
	t, err := g.getTopic(name)
	if err != nil {
		return err
	}

	return t.Delete(context.Background())
}
