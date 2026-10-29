package rediscluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAwsElastiCacheParameterGroupFamily(t *testing.T) {
	// engineVersion is constrained by the CRD enum to 6.x / 7.0 / 7.1; anything else falls back to "".
	cases := []struct {
		engineVersion string
		expected      string
	}{
		{"6.x", "redis6.x"},
		{"7.0", "redis7"},
		{"7.1", "redis7"},
		{"5.0", ""},
		{"8.0", ""},
		{"unknown", ""},
		{"", ""},
	}

	for _, c := range cases {
		t.Run(c.engineVersion, func(t *testing.T) {
			assert.Equal(t, c.expected, GetAwsElastiCacheParameterGroupFamily(c.engineVersion))
		})
	}
}
