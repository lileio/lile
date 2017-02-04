package lile

import (
	"testing"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/stretchr/testify/assert"
)

func TestBlankNewServer(t *testing.T) {
	s := NewServer()
	assert.NotNil(t, s)
}

func TestName(t *testing.T) {
	s := NewServer(
		Name("somethingcool"),
	)
	assert.NotNil(t, s)
	assert.Equal(t, "somethingcool", s.opts.name)
}

func TestPort(t *testing.T) {
	s := NewServer(
		Port(":9999"),
	)
	assert.NotNil(t, s)
	assert.Equal(t, ":9999", s.opts.port)
}

func TestPrometheusEnabled(t *testing.T) {
	s := NewServer(
		PrometheusEnabled(false),
	)
	assert.NotNil(t, s)
	assert.False(t, s.opts.prometheus)
}

func TestPrometheusPort(t *testing.T) {
	s := NewServer(
		PrometheusPort(":4321"),
	)
	assert.NotNil(t, s)
	assert.Equal(t, ":4321", s.opts.prometheusPort)
}

func TestPrometheusAddr(t *testing.T) {
	s := NewServer(
		PrometheusAddr("/prom"),
	)
	assert.NotNil(t, s)
	assert.Equal(t, "/prom", s.opts.prometheusAddr)
}

func TestUnary(t *testing.T) {
	s := NewServer(
		AddUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
		),
	)
	assert.NotNil(t, s)
	assert.NotEmpty(t, s.opts.unaryInts)
}

func TestStream(t *testing.T) {
	s := NewServer(
		AddStreamInterceptor(
			grpc_prometheus.StreamServerInterceptor,
		),
	)
	assert.NotNil(t, s)
	assert.NotEmpty(t, s.opts.streamInts)
}

type someImpl struct {
}

func TestTracingEnabled(t *testing.T) {
	s := NewServer(
		TracingEnabled(false),
	)
	assert.NotNil(t, s)
	assert.False(t, s.opts.tracing)
}

func TestTracing(t *testing.T) {
	c, err := zipkin.NewHTTPCollector("http://zipkin/")
	z, err := zipkin.NewTracer(zipkin.NewRecorder(c, false, "", ""))

	s := NewServer(
		Tracer(z),
	)

	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.NotNil(t, s.opts.tracer)
}

func TestImplementation(t *testing.T) {
	s := NewServer(
		Implementation(func(g *grpc.Server) {}),
	)
	assert.NotNil(t, s)
	assert.NotEmpty(t, s.opts.implementation)
}

func TestListenAndServe(t *testing.T) {
	s := NewServer()
	go s.ListenAndServe()
}
