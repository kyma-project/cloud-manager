package feature

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemovedFlagAbsent is a repo-wide guard that fails if any .go file defines or
// references the removed "Api"+"Disabled" symbol, or any .yaml/.yml file carries an
// "api"+"Disabled:" flag block. String literals are split to avoid self-match.
func TestRemovedFlagAbsent(t *testing.T) {
	goSymbol := "Api" + "Disabled"
	yamlKey := "api" + "Disabled:"

	moduleRoot := findModuleRoot(t)

	var goViolations []string
	var yamlViolations []string

	err := filepath.WalkDir(moduleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			// Skip vendor, hidden dirs (.git, .claude, …), and generated dirs that are not part of the source tree.
			if name == "vendor" || name == "bin" || name == "testbin" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if strings.Contains(string(data), goSymbol) {
				rel, _ := filepath.Rel(moduleRoot, path)
				goViolations = append(goViolations, rel)
			}
		case ".yaml", ".yml":
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, line := range strings.Split(string(data), "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, yamlKey) {
					rel, _ := filepath.Rel(moduleRoot, path)
					yamlViolations = append(yamlViolations, rel)
					break
				}
			}
		}
		return nil
	})
	assert.NoError(t, err, "repo walk must not fail")
	assert.Empty(t, goViolations,
		"found Go files referencing the removed %s symbol — delete or rewrite them", goSymbol)
	assert.Empty(t, yamlViolations,
		"found YAML files with an apiDisabled: block — remove or convert them")
}

// findModuleRoot resolves the module root from the package directory. `go test` sets the
// working directory to the package under test (pkg/feature), so two levels up is the module
// root. This avoids runtime.Caller, whose embedded path is module-relative under -trimpath.
func findModuleRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err, "resolving module root from package directory")
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("findModuleRoot: no go.mod at resolved root %s — test cwd assumption broke", root)
	}
	return root
}
