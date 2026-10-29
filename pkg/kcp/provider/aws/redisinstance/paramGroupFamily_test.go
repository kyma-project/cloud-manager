package redisinstance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAwsElastiCacheParameterGroupFamily(t *testing.T) {
	cases := []struct {
		engineVersion string
		expected      string
	}{
		{"2.6.13", "redis2.6"},
		{"2.8.24", "redis2.8"},
		{"3.2.6", "redis3.2"},
		{"4.0.10", "redis4.0"},
		{"5.0.6", "redis5.0"},
		{"6.0", "redis6.x"},
		{"6.2", "redis6.x"},
		{"7.0", "redis7"},
		{"7.1", "redis7"},
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
