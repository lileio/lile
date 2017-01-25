package lile

import (
	"testing"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
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
