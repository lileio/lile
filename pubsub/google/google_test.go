package google

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/lileio/lile/pubsub"
	"github.com/lileio/lile/test"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	zipkintracer "github.com/openzipkin/zipkin-go-opentracing"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestGooglePublishSubscribe(t *testing.T) {
	key := os.Getenv("GOOGLE_KEY")
	if key != "" {
		err := ioutil.WriteFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), []byte(key), os.ModePerm)
		if err != nil {
			panic(err)
		}

	}

	if os.Getenv("GOOGLE_PUBSUB_PROJECT_ID") == "" {
		assert.Fail(t, "No GOOGLE_PUBSUB_PROJECT_ID is set")
		return
	}

	sub := "lile_" + uuid.NewV1().String()
	ps, err := NewGoogleCloud(os.Getenv("GOOGLE_PUBSUB_PROJECT_ID"), sub)
	assert.Nil(t, err)
	assert.NotNil(t, ps)

	done := make(chan bool)

	topic := "lile_topic"
	_, err = ps.getTopic(topic)
	assert.Nil(t, err)

	a := test.Account{Name: "Alex"}
	ps.subscribe(topic, func(ctx context.Context, m pubsub.Msg) error {
		assert.NotNil(t, opentracing.SpanFromContext(ctx))

		var ac test.Account
		proto.Unmarshal(m.Data, &ac)

		assert.Equal(t, ac.Name, a.Name)

		done <- true
		return nil
	}, 10*time.Second, true, done)

	// Wait for subscription
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		assert.Fail(t, "Subscription failed after timeout")
	}

	const op = "test_event"
	recorder := zipkintracer.NewInMemoryRecorder()
	tracer, err := zipkintracer.NewTracer(
		recorder,
		zipkintracer.ClientServerSameSpan(true),
		zipkintracer.DebugMode(true),
		zipkintracer.TraceID128Bit(true),
	)
	opentracing.SetGlobalTracer(tracer)
	if err != nil {
		t.Fatalf("Unable to create Tracer: %+v", err)
	}

	span := tracer.StartSpan(op)
	span.LogFields(
		log.String("event", "soft error"),
		log.String("sql", "select * from something;"),
		log.Int("waited.millis", 1500))
	span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	err = ps.Publish(ctx, topic, &a)
	assert.Nil(t, err)

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		assert.Fail(t, "Subscribe failed after timeout")
	}

	assert.Nil(t, ps.deleteTopic(topic))
}
