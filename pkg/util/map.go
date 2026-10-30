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

	if len(templatesMap) == 0 {
		return result
	}
	if dataMap == nil {
		dataMap = map[string]string{}
	}

	// Build a single regex matching {{[-\s]*.key[-\s]*}} for each key in dataMap.
	// This mimics text/template field access ({{.key}}, {{ .key }}, {{- .key -}}, etc.)
	// without executing arbitrary template actions like range, call, pipe, etc.
	var re *regexp.Regexp
	if len(dataMap) > 0 {
		keyPatterns := make([]string, 0, len(dataMap))
		for k := range dataMap {
			keyPatterns = append(keyPatterns, regexp.QuoteMeta(k))
		}
		// Capture group 1 contains the exact matched key — used directly in the closure below.
		// Key order within the alternation does not matter: Go's RE2 engine requires the full
		// pattern (including \s*-?\}\}) to match, so a short key like "h" cannot steal the match
		// from "host" — "ost\s*-?\}\}" won't follow "h" in "{{.host}}".
		joined := "(" + strings.Join(keyPatterns, "|") + ")"
		re = regexp.MustCompile(`\{\{-?\s*\.` + joined + `\s*-?\}\}`)
	}

	for k, v := range templatesMap {
		if re == nil {
			result[k] = []byte(v)
			continue
		}
		result[k] = re.ReplaceAllFunc([]byte(v), func(match []byte) []byte {
			key := re.FindSubmatch(match)[1]
			if val, ok := dataMap[string(key)]; ok {
				return []byte(val)
			}
			return match
		})
	}

	return result
}
