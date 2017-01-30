package lile

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func tracerFromEnv(opts options) opentracing.Tracer {
	var c zipkin.Collector
	var err error

	w := logrus.StandardLogger().Writer()
	l := log.New(w, "Zipkin: ", 0)

	if u := os.Getenv("ZIPKIN_HTTP_ENDPOINT"); u != "" {
		c, err = zipkin.NewHTTPCollector(
			u,
			zipkin.HTTPLogger(zipkin.LogWrapper(l)),
		)
	}

	if u := os.Getenv("ZIPKIN_SCRIBE_ENDPOINT"); u != "" {
		c, err = zipkin.NewScribeCollector(
			u, time.Second*10,
			zipkin.ScribeLogger(zipkin.LogWrapper(l)),
		)
	}

	if u := os.Getenv("ZIPKIN_KAFKA_ENDPOINTS"); u != "" {
		c, err = zipkin.NewKafkaCollector(
			strings.Split(u, ","),
			zipkin.KafkaLogger(zipkin.LogWrapper(l)),
		)
	}

	if err != nil {
		logrus.Fatal(err)
	}

	t, err := zipkinTracer(opts, c)
	if err != nil {
		logrus.Fatal(err)
	}

	return t
}

func zipkinTracer(opts options, collector zipkin.Collector) (t opentracing.Tracer, err error) {
	return zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, opts.name, opts.name),
		zipkin.ClientServerSameSpan(true), // for Zipkin V1 RPC span style
	)
}
