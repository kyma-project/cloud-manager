package composed_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// forbiddenPatterns lists variable references that must never appear verbatim in SKR
// updateStatus files — each carries raw KCP-internal error text that must be sanitized
// before being surfaced to tenants. Ready/Updating messages are intentionally forwarded
// verbatim and are therefore not listed here.
var forbiddenPatterns = []string{
	"kcpCondErr.Message",
}

// TestNoVerbatimKcpConditionMessageInSKR asserts that no file under pkg/skr/
// contains a verbatim copy of a KCP condition message.
// All SKR writers must use composed.SanitizeErrorForTenant(kcpCondErr.Reason) instead.
func TestNoVerbatimKcpConditionMessageInSKR(t *testing.T) {
	// os.Getwd() in go test returns the package source directory — unaffected by -trimpath.
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not determine working directory: %v", err)
	}
	// pkgDir = .../pkg/composed — go up one level to pkg/, then into skr/
	skrRoot := filepath.Join(filepath.Dir(pkgDir), "skr")

	if _, statErr := os.Stat(skrRoot); os.IsNotExist(statErr) {
		t.Fatalf("pkg/skr directory not found at %s — test must run with source tree present", skrRoot)
	}

	err = filepath.WalkDir(skrRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		src := string(content)
		for _, pattern := range forbiddenPatterns {
			if strings.Contains(src, pattern) {
				t.Errorf("forbidden pattern %q found in %s — use composed.SanitizeErrorForTenant(...Reason) instead", pattern, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("error walking pkg/skr: %v", err)
	}
}
