package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocation(t *testing.T) {
	t.Run("uses spec location when set", func(t *testing.T) {
		result := getLocation("europe-west3", "us-central1")
		assert.Equal(t, "europe-west3", result)
	})

	t.Run("falls back to scope region when spec location is empty", func(t *testing.T) {
		result := getLocation("", "us-central1")
		assert.Equal(t, "us-central1", result)
	})
}
