package util

import (
	"maps"
	"regexp"
	"strings"
)

// MergeMaps merges two maps of the same key-value types.
// Does not modify maps. Returns new map as output.
// If onCollisionOverwrite is true, collision is resolved with second map value for key. Otherwise, first map value for key is used.
func MergeMaps[K comparable, V any](first, second map[K]V, onCollisionOverwrite bool) map[K]V {
	if first == nil {
		first = make(map[K]V)
	}
	if second == nil {
		second = make(map[K]V)
	}
	result := make(map[K]V)

	maps.Copy(result, first)

	for k, v := range second {
		if _, exists := result[k]; !exists || onCollisionOverwrite {
			result[k] = v
		}
	}

	return result
}

// ParseTemplatesMapToBytesMap interpolates values using {{.key}} syntax against dataMap keys.
func ParseTemplatesMapToBytesMap(templatesMap, dataMap map[string]string) map[string][]byte {
	result := map[string][]byte{}

	if templatesMap == nil {
		return result
	}
	if dataMap == nil {
		dataMap = map[string]string{}
	}

	// Build a single regex matching {{[-\s]*.key[-\s]*}} for each key in dataMap.
	// This mimics text/template field access ({{.key}}, {{ .key }}, {{- .key -}}, etc.)
	// without executing arbitrary template actions like range, call, pipe, etc.
	keyPatterns := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keyPatterns = append(keyPatterns, regexp.QuoteMeta(k))
	}
	var re *regexp.Regexp
	if len(keyPatterns) > 0 {
		// Pattern: optional trim marker (-), optional whitespace, .key, optional whitespace, optional trim marker (-)
		joined := "(" + strings.Join(keyPatterns, "|") + ")"
		re = regexp.MustCompile(`\{\{-?\s*\.` + joined + `\s*-?\}\}`)
	}

	for k, v := range templatesMap {
		if re == nil {
			result[k] = []byte(v)
			continue
		}
		result[k] = re.ReplaceAllFunc([]byte(v), func(match []byte) []byte {
			sub := re.FindSubmatch(match)
			if len(sub) < 2 {
				return match
			}
			return []byte(dataMap[string(sub[1])])
		})
	}

	return result
}
