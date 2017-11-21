// package fromenv provides utilities to create lile options from environment
// variables. fromenv will error with fatal if it cannot resolve or errors
package fromenv

import (
	"fmt"
	"os"

	"github.com/lileio/pubsub"
	"github.com/lileio/pubsub/google"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/sirupsen/logrus"
)

func Tracer(name string) opentracing.Tracer {
	var collector zipkin.Collector
	var err error

	zipkin_host := os.Getenv("ZIPKIN_SERVICE_HOST")
	if zipkin_host != "" {
		addr := fmt.Sprintf("http://%s:%s/api/v1/spans",
			os.Getenv("ZIPKIN_SERVICE_HOST"),
			os.Getenv("ZIPKIN_SERVICE_PORT"))
		collector, err = zipkin.NewHTTPCollector(addr)
		if err != nil {
			logrus.Fatalf("unable to create Zipkin HTTP collector: %+v", err)
		}

		logrus.Infof("Using Zipkin HTTP tracer: %s", addr)
	}

	if collector == nil {
		logrus.Infof("Using Zipkin Global tracer")
		return opentracing.GlobalTracer()
	}

	// create recorder.
	recorder := zipkin.NewRecorder(collector, false, "", name)

	// create tracer.
	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(true),
	)
	if err != nil {
		logrus.Fatalf("unable to create Zipkin tracer: %+v", err)
	}

	// explicitly set our tracer to be the default tracer.
	opentracing.InitGlobalTracer(tracer)

	return tracer
}

func PubSubProvider() pubsub.Provider {
	gpid := os.Getenv("GOOGLE_PUBSUB_PROJECT_ID")
	if gpid != "" {
		gc, err := google.NewGoogleCloud(gpid)
		if err != nil {
			logrus.Fatalf("fronenv: Google Cloud pubsub err: %s", err)
			return nil
		}

		logrus.Infof("Using Google Cloud pubsub: %s", gpid)
		return gc
	}

	logrus.Warn("Using noop pubsub provider")
	return pubsub.NoopProvider{}
}
