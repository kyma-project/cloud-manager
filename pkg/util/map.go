package util

import (
	"maps"
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

	pairs := make([]string, 0, len(dataMap)*2)
	for k, v := range dataMap {
		pairs = append(pairs, "{{."+k+"}}", v)
	}
	r := strings.NewReplacer(pairs...)

	for k, v := range templatesMap {
		result[k] = []byte(r.Replace(v))
	}

	return result
}
