package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	t.Run("MergeMaps", func(t *testing.T) {

		t.Run("should return expected map", func(t *testing.T) {
			first := map[string]string{
				"foo": "bar",
			}
			second := map[string]string{
				"baz": "qux",
			}

			result := MergeMaps(first, second, false)

			assert.Equal(t, 2, len(result))
			assert.Equal(t, "bar", result["foo"])
			assert.Equal(t, "qux", result["baz"])
		})

		t.Run("should not modify original maps", func(t *testing.T) {
			first := map[string]string{
				"foo": "bar",
				"rab": "bar",
			}
			second := map[string]string{
				"foo": "oof",
				"baz": "qux",
			}

			MergeMaps(first, second, true)

			assert.Equal(t, 2, len(first))
			assert.Equal(t, 2, len(first))
			assert.Equal(t, "bar", first["foo"])
			assert.Equal(t, "bar", first["rab"])
			assert.Equal(t, "oof", second["foo"])
			assert.Equal(t, "qux", second["baz"])
		})

		t.Run("should handle collission (onCollisionOverwrite=false)", func(t *testing.T) {
			first := map[string]string{
				"foo": "bar",
			}
			second := map[string]string{
				"foo": "baz",
			}

			result := MergeMaps(first, second, false)

			assert.Equal(t, 1, len(result))
			assert.Equal(t, "bar", result["foo"])
		})

		t.Run("should handle collission (onCollisionOverwrite=true)", func(t *testing.T) {
			first := map[string]string{
				"foo": "bar",
			}
			second := map[string]string{
				"foo": "baz",
			}

			result := MergeMaps(first, second, true)

			assert.Equal(t, 1, len(result))
			assert.Equal(t, "baz", result["foo"])
		})

		t.Run("should handle empty maps", func(t *testing.T) {
			first := map[string]string{}
			second := map[string]string{}

			result := MergeMaps(first, second, false)

			assert.Equal(t, 0, len(result))
		})

		t.Run("should handle first map nil", func(t *testing.T) {
			theMap := map[string]string{
				"foo": "bar",
			}

			result := MergeMaps(theMap, nil, false)

			assert.Equal(t, 1, len(result))
			assert.Equal(t, "bar", result["foo"])
		})

		t.Run("should handle second map nil", func(t *testing.T) {
			theMap := map[string]string{
				"foo": "bar",
			}

			result := MergeMaps(nil, theMap, false)

			assert.Equal(t, 1, len(result))
			assert.Equal(t, "bar", result["foo"])
		})
	})
}

func TestParseTemplatesMapToBytesMap(t *testing.T) {
	t.Run("ParseTemplatesMapToBytesMap", func(t *testing.T) {

		t.Run("should return expected map", func(t *testing.T) {
			templates := map[string]string{
				"greeting1": "Hello {{.rank}} {{.name}}.",
				"greeting2": "Hello.",
				"greeting3": "Hello. {{.unknown}}.",
			}
			data := map[string]string{
				"rank": "major",
				"name": "laser",
			}

			result := ParseTemplatesMapToBytesMap(templates, data)

			assert.Equal(t, 3, len(result))
			assert.Equal(t, "Hello major laser.", string(result["greeting1"]))
			assert.Equal(t, "Hello.", string(result["greeting2"]))
			assert.Equal(t, "Hello. <no value>.", string(result["greeting3"]))
		})

		t.Run("should handle empty template map", func(t *testing.T) {
			data := map[string]string{
				"rank": "major",
				"name": "laser",
			}

			result := ParseTemplatesMapToBytesMap(nil, data)

			assert.Equal(t, 0, len(result))
		})

		t.Run("should handle empty data map", func(t *testing.T) {
			templates := map[string]string{
				"greeting1": "Hello {{.rank}} {{.name}}.",
				"greeting2": "Hello.",
				"greeting3": "Hello. {{.unknown}}.",
			}

			result := ParseTemplatesMapToBytesMap(templates, nil)

			assert.Equal(t, 3, len(result))
			assert.Equal(t, string("Hello <no value> <no value>."), string(result["greeting1"]))
			assert.Equal(t, string("Hello."), string(result["greeting2"]))
			assert.Equal(t, string("Hello. <no value>."), string(result["greeting3"]))
		})

	})
}
