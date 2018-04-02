package lile_test

import (
	"testing"

	"github.com/lileio/lile"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	lile.NewService("test_service")
	assert.Equal(t, lile.GlobalService().Name, "test_service")
}
