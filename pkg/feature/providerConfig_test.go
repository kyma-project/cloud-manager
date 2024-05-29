package feature

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviderConfig(t *testing.T) {
	env := abstractions.NewMockedEnvironment(map[string]string{
		"foo":                "foo",
		"FOO":                "FOO",
		"FF_SOME_BOOL_TRUE":  "true",
		"FF_SOME_BOOL_FALSE": "false",
		"FF_SOME_STRING":     "abc",
		"FF_SOME_INT":        "123",
		"FF_SOME_FLOAT":      "2.5",
		"FF_ARR":             `[1,"b"]`,
		"FF_OBJ":             `{"a": 1, "b": "bbb"}`,
	})
	p := NewProviderConfig(env)

	ctx := context.Background()

	t.Run("bool true", func(t *testing.T) {
		assert.Equal(t, true, p.BoolVariation(ctx, "someBoolTrue", false))
	})
	t.Run("bool false", func(t *testing.T) {
		assert.Equal(t, false, p.BoolVariation(ctx, "someBoolFalse", true))
	})
	t.Run("string", func(t *testing.T) {
		assert.Equal(t, "abc", p.StringVariation(ctx, "someString", ""))
	})
	t.Run("int", func(t *testing.T) {
		assert.Equal(t, 123, p.IntVariation(ctx, "someInt", 999))
	})
	t.Run("float", func(t *testing.T) {
		assert.InDelta(t, 2.5, p.Float64Variation(ctx, "someFloat", 0.1), 0.01)
	})
	t.Run("arr", func(t *testing.T) {
		arr := p.JSONArrayVariation(ctx, "arr", nil)
		expected := []interface{}{1, "b"}
		assert.Equal(t, fmt.Sprintf("%#v", expected), fmt.Sprintf("%#v", arr))
	})
	t.Run("map", func(t *testing.T) {
		obj := p.JSONVariation(ctx, "obj", nil)
		expected := map[string]interface{}{
			"a": 1,
			"b": "bbb",
		}
		assert.Equal(t, fmt.Sprintf("%#v", expected), fmt.Sprintf("%#v", obj))
	})
}
