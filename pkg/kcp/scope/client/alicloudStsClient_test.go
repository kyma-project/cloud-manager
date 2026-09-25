package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The darabonba-openapi CallApi wraps a json response under a "body" key. Guards
// the bug where reading AccountId at the top level always failed.
func TestAccountIdFromCallApiBody(t *testing.T) {
	wrapped := map[string]any{
		"body":       map[string]any{"AccountId": "196813200012", "Arn": "acs:ram::196813200012:user/cloud-manager"},
		"headers":    map[string]any{},
		"statusCode": 200,
	}
	id, ok := accountIdFromCallApiBody(wrapped)
	assert.True(t, ok)
	assert.Equal(t, "196813200012", id)

	// Defensive flat shape.
	id, ok = accountIdFromCallApiBody(map[string]any{"AccountId": "123"})
	assert.True(t, ok)
	assert.Equal(t, "123", id)

	// Missing.
	_, ok = accountIdFromCallApiBody(map[string]any{"body": map[string]any{}})
	assert.False(t, ok)
}
