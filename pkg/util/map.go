package util

import (
	"bytes"
	"text/template"
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

	for k, v := range first {
		result[k] = v
	}

	for k, v := range second {
		if _, exists := result[k]; !exists || onCollisionOverwrite {
			result[k] = v
		}
	}

	return result
}

// Templates format map[k]="Hello {{.name}}"
func ParseTemplatesMapToBytesMap(templatesMap, dataMap map[string]string) map[string][]byte {
	result := map[string][]byte{}

	if templatesMap == nil {
		return result
	}
	if dataMap == nil {
		dataMap = map[string]string{}
	}

	for k, v := range templatesMap {
		var parseResult bytes.Buffer
		template, err := template.New(k).Parse(v)
		if err != nil {
			result[k] = []byte(v)
			continue
		}
		err = template.Execute(&parseResult, dataMap)
		if err != nil {
			result[k] = []byte(v)
			continue
		}
		result[k] = parseResult.Bytes()
	}

	return result
}

func AreByteMapsEqual(first, second map[string][]byte) bool {
	if len(first) != len(second) {
		return false
	}

	for key, firstValue := range first {
		secondValue, exists := second[key]
		if !exists {
			return false
		}

		if !bytes.Equal(firstValue, secondValue) {
			return false
		}
	}

	return true
}

func AreStringMapsEqual(first, second map[string]string) bool {
	if len(first) != len(second) {
		return false
	}

	for key, value1 := range first {
		value2, exists := second[key]
		if !exists || value1 != value2 {
			return false
		}
	}

	return true
}
