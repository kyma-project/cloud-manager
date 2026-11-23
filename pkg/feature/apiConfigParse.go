package feature

import "gopkg.in/yaml.v3"

// parseConfigBytes unmarshals raw YAML config bytes into a root map.
// It returns an error on parse failure so callers can distinguish malformed
// from empty configs.
func parseConfigBytes(b []byte) (map[string]any, error) {
	var root map[string]any
	if err := yaml.Unmarshal(b, &root); err != nil {
		return nil, err
	}
	return root, nil
}

// blockHasContent reports whether a flag block carries targeting rules or a defaultRule.
// An empty block is not treated as an inconsistency.
func blockHasContent(block any) bool {
	m, ok := block.(map[string]any)
	if !ok {
		return false
	}
	if t, ok := m["targeting"].([]any); ok && len(t) > 0 {
		return true
	}
	_, hasDefault := m["defaultRule"]
	return hasDefault
}
