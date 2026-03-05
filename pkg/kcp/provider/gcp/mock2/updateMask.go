package mock2

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// UpdateMask updates fields in instance from update based on the updateMask.
// Only fields listed in updateMask.Paths are copied from update to instance.
// Field names in updateMask match the json tag field names and can contain
// nested paths separated by periods (e.g., "field.child.destination").
func UpdateMask[T protoadapt.MessageV1](instance T, update T, updateMask *fieldmaskpb.FieldMask) error {
	if updateMask == nil || len(updateMask.Paths) == 0 {
		return nil
	}

	// Convert instance to protov2 message
	instanceV2 := protoadapt.MessageV2Of(instance)
	updateV2 := protoadapt.MessageV2Of(update)

	// Marshal both to JSON as unstructured maps
	marshaler := protojson.MarshalOptions{EmitUnpopulated: false}

	instanceBytes, err := marshaler.Marshal(instanceV2)
	if err != nil {
		return fmt.Errorf("failed to marshal instance to JSON: %w", err)
	}

	updateBytes, err := marshaler.Marshal(updateV2)
	if err != nil {
		return fmt.Errorf("failed to marshal update to JSON: %w", err)
	}

	// Parse JSON into unstructured maps
	instanceMap, err := parseJSON(instanceBytes)
	if err != nil {
		return fmt.Errorf("failed to parse instance JSON: %w", err)
	}

	updateMap, err := parseJSON(updateBytes)
	if err != nil {
		return fmt.Errorf("failed to parse update JSON: %w", err)
	}

	// Apply each path from updateMask
	for _, path := range updateMask.Paths {
		value, found := getValueAtPath(updateMap, path)
		if found {
			setValueAtPath(instanceMap, path, value)
		}
	}

	// Marshal the modified instance map back to JSON
	modifiedBytes, err := marshalJSON(instanceMap)
	if err != nil {
		return fmt.Errorf("failed to marshal modified instance: %w", err)
	}

	// Unmarshal back into the instance
	// First reset the instance to clear all fields
	proto.Reset(instanceV2)

	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaler.Unmarshal(modifiedBytes, instanceV2); err != nil {
		return fmt.Errorf("failed to unmarshal modified JSON to instance: %w", err)
	}

	return nil
}

// parseJSON parses JSON bytes into an unstructured map
func parseJSON(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := jsonUnmarshal(data, &result); err != nil {
		return nil, err
	}
	if result == nil {
		result = make(map[string]any)
	}
	return result, nil
}

// marshalJSON marshals an unstructured map to JSON bytes
func marshalJSON(data map[string]any) ([]byte, error) {
	return jsonMarshal(data)
}

// getValueAtPath retrieves a value at a dot-separated path from a nested map
func getValueAtPath(data map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	current := any(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// setValueAtPath sets a value at a dot-separated path in a nested map
// Creates intermediate maps as needed
func setValueAtPath(data map[string]any, path string, value any) {
	parts := strings.Split(path, ".")

	if len(parts) == 1 {
		data[path] = value
		return
	}

	current := data
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			// Create intermediate map
			newMap := make(map[string]any)
			current[part] = newMap
			current = newMap
		} else {
			nextMap, ok := next.(map[string]any)
			if !ok {
				// Cannot traverse non-map, create new map
				newMap := make(map[string]any)
				current[part] = newMap
				current = newMap
			} else {
				current = nextMap
			}
		}
	}

	current[parts[len(parts)-1]] = value
}

// jsonUnmarshal unmarshals JSON data into v
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// jsonMarshal marshals v to JSON
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
