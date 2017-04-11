package pubsub

import (
	"context"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	ctxNet "golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/jpillora/backoff"
	"github.com/sirupsen/logrus"
)

var mutex = &sync.Mutex{}

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

func (g *GoogleCloud) getTopic(name string) (*pubsub.Topic, error) {
	if g.topics[name] != nil {
		return g.topics[name], nil
	}

	ctx := ctxNet.Background()
	topic := g.client.Topic(name)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if ok {
		return topic, nil
	}

	t, err := g.client.CreateTopic(ctx, name)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (g *GoogleCloud) Publish(ctx context.Context, topic string, msg proto.Message) error {
	b, err := proto.Marshal(msg)
	if err != nil {
		logrus.Errorf("Cant marshal msg for topic %s, err: %v", topic, err)
	}

	mutex.Lock()
	t, err := g.getTopic(topic)
	mutex.Unlock()
	if err != nil {
		return err
	}

	cb := ctxNet.Background()
	res := t.Publish(cb, &pubsub.Message{
		Data: b,
	})

	_, err = res.Get(cb)
	return err
}

func (g *GoogleCloud) Subscribe(topic string, h MsgHandler, deadline time.Duration, autoAck bool) {
	go func() {
		var sub *pubsub.Subscription
		var err error
		c := ctxNet.Background()
		b := &backoff.Backoff{
			Min: 500 * time.Millisecond,
			Max: 10 * time.Second,
		}

		// Subscribe with backoff for failure (i.e topic doesn't exist yet)
		for {
			t := g.client.Topic(topic)
			subName := g.subName + "--" + topic
			sub, err = g.client.CreateSubscription(c, subName, t, deadline, nil)
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

		// Listen to messages and call the MsgHandler
		for {
			err = sub.Receive(c, func(ctx ctxNet.Context, m *pubsub.Message) {
				logrus.Infof("Recevied on topic %s, id: %s", topic, m.ID)

				msg := Msg{
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

				err := h(ctx, msg)
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
