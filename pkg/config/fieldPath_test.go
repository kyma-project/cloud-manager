package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConcatFieldPath(t *testing.T) {
	ab := "aaa.bbb"
	cd := "ccc.ddd"

	assert.Equal(t, "aaa.bbb.ccc.ddd", ConcatFieldPath(ab, cd))
}
