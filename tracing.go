package lile

import (
	"log"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	jaeger "github.com/uber/jaeger-client-go"
	jz "github.com/uber/jaeger-client-go/transport/zipkin"
)

func tracerFromEnv(opts options) *opentracing.Tracer {
	if u := os.Getenv("ZIPKIN_HTTP_ENDPOINT"); u != "" {
		transport, err := jz.NewHTTPTransport(
			u,
			jz.HTTPBatchSize(1),
			jz.HTTPLogger(jaeger.StdLogger),
		)

		if err != nil {
			logrus.Printf("Zipkin connection error: %s", err)
		}

		tracer, _ := jaeger.NewTracer(
			opts.name,
			jaeger.NewConstSampler(true),
			jaeger.NewRemoteReporter(transport,
				jaeger.ReporterOptions.Logger(jaeger.StdLogger),
			),
		)

		logrus.Printf("Zipkin: using HTTP collector at %s", u)
		return &tracer
	}

	if u := os.Getenv("ZIPKIN_KAFKA_ENDPOINTS"); u != "" {
		w := logrus.StandardLogger().Writer()
		l := log.New(w, "Zipkin: ", 0)

		c, err := zipkin.NewKafkaCollector(
			strings.Split(u, ","),
			zipkin.KafkaLogger(zipkin.LogWrapper(l)),
		)

		if err != nil {
			logrus.Printf("Zipkin connection error: %s", err)
		}

		t, err := zipkin.NewTracer(
			zipkin.NewRecorder(c, false, opts.name, opts.name),
			zipkin.ClientServerSameSpan(true),
		)

		if err != nil {
			logrus.Printf("Zipkin tracer error: %s", err)
		}

		return &t
	}

	return nil
}

func zipkinTracer(opts options) *opentracing.Tracer {
	return tracerFromEnv(opts)
}
