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
				"printf":    `{{printf "%s" "exec"}}`,
				"call":      `{{call .fn}}`,
				"range":     `{{range .items}}x{{end}}`,
				"pipe":      `{{.name | html}}`,
				"define":    `{{define "t"}}x{{end}}`,
			}
			data := map[string]string{
				"rank": "major",
				"name": "laser",
			}

			result := ParseTemplatesMapToBytesMap(templates, data)

			assert.Equal(t, 8, len(result))
			assert.Equal(t, "Hello major laser.", string(result["greeting1"]))
			assert.Equal(t, "Hello.", string(result["greeting2"]))
			assert.Equal(t, "Hello. {{.unknown}}.", string(result["greeting3"]))
			assert.Equal(t, `{{printf "%s" "exec"}}`, string(result["printf"]))
			assert.Equal(t, `{{call .fn}}`, string(result["call"]))
			assert.Equal(t, `{{range .items}}x{{end}}`, string(result["range"]))
			assert.Equal(t, `{{.name | html}}`, string(result["pipe"]))
			assert.Equal(t, `{{define "t"}}x{{end}}`, string(result["define"]))

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
			assert.Equal(t, string("Hello {{.rank}} {{.name}}."), string(result["greeting1"]))
			assert.Equal(t, string("Hello."), string(result["greeting2"]))
			assert.Equal(t, string("Hello. {{.unknown}}."), string(result["greeting3"]))
		})

		t.Run("should handle whitespace variants around key", func(t *testing.T) {
			templates := map[string]string{
				"nospace":         "{{.name}}",
				"one_space":       "{{ .name }}",
				"multi_space":     "{{  .name  }}",
				"tab":             "{{\t.name\t}}", // tab is matched by \s*; text/template would parse-error on this
				"trim_marker":     "{{- .name -}}",
				"trim_left_only":  "{{- .name }}",
				"trim_right_only": "{{ .name -}}",
			}
			data := map[string]string{"name": "laser"}

			result := ParseTemplatesMapToBytesMap(templates, data)

			assert.Equal(t, "laser", string(result["nospace"]))
			assert.Equal(t, "laser", string(result["one_space"]))
			assert.Equal(t, "laser", string(result["multi_space"]))
			assert.Equal(t, "laser", string(result["tab"]))
			assert.Equal(t, "laser", string(result["trim_marker"]))
			assert.Equal(t, "laser", string(result["trim_left_only"]))
			assert.Equal(t, "laser", string(result["trim_right_only"]))
		})

		t.Run("should substitute repeated occurrences of the same key", func(t *testing.T) {
			templates := map[string]string{
				"repeated": "{{.name}} and {{.name}}",
			}
			data := map[string]string{"name": "laser"}

			result := ParseTemplatesMapToBytesMap(templates, data)

			assert.Equal(t, "laser and laser", string(result["repeated"]))
		})

		t.Run("should not match a prefix key when a longer key is present", func(t *testing.T) {
			// dataMap has both "h" and "host"; {{.host}} must resolve to "example.com", not "laser" + "ost"
			templates := map[string]string{
				"full":   "{{.host}}",
				"prefix": "{{.h}}",
			}
			data := map[string]string{
				"h":    "laser",
				"host": "example.com",
			}

			result := ParseTemplatesMapToBytesMap(templates, data)

			assert.Equal(t, "example.com", string(result["full"]))
			assert.Equal(t, "laser", string(result["prefix"]))
		})

		t.Run("should not execute CVE DoS payload (deeply nested range)", func(t *testing.T) {
			// Regression test: the original vulnerability allowed a tenant to supply
			// deeply nested {{range}} directives that exhausted CPU/memory in the shared pod.
			dos := `{{range $i, $v := .}}{{range $j, $w := .}}{{range $k, $u := .}}{{end}}{{end}}{{end}}`
			templates := map[string]string{"x": dos}
			data := map[string]string{"name": "laser"}

			result := ParseTemplatesMapToBytesMap(templates, data)

			// The payload must pass through verbatim — not executed, not modified.
			assert.Equal(t, dos, string(result["x"]))
		})

	})
}
