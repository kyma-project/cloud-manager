package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunDivergentConfig asserts that a divergent fixture causes a non-zero exit
// code, a FAIL: summary line, and at least one DIVERGE detail line on stdout.
func TestRunDivergentConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"ffparity", "../../pkg/feature/testdata/apiParityDivergent.yaml"}, &stdout, &stderr)

	assert.NotEqual(t, 0, code, "divergent config must exit non-zero")
	assert.Contains(t, stdout.String(), "FAIL:", "stdout must contain FAIL: summary line")
	assert.Contains(t, stdout.String(), "DIVERGE ", "stdout must contain at least one DIVERGE detail line")
	assert.Contains(t, stdout.String(), "feature=", "DIVERGE detail line must carry structured fields (feature=...)")
}

// TestRunNonexistentPath asserts that a missing config exits non-zero and does not
// print PASS or any zero-divergence indicator (AC4: missing file must error).
func TestRunNonexistentPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"ffparity", "/does/not/exist/ffparity_test_missing.yaml"}, &stdout, &stderr)

	assert.NotEqual(t, 0, code, "missing config must exit non-zero")
	assert.NotContains(t, stdout.String(), "PASS", "stdout must not contain PASS for a missing config")
	assert.NotEmpty(t, stderr.String(), "stderr must contain an error message")
}

// TestRunMalformedConfig asserts that a malformed YAML config exits non-zero and
// produces a parse-failure error (AC4: malformed config must error, not zero-divergence pass).
func TestRunMalformedConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"ffparity", "../../pkg/feature/testdata/apiMalformed.yaml"}, &stdout, &stderr)

	assert.NotEqual(t, 0, code, "malformed config must exit non-zero")
	assert.NotContains(t, stdout.String(), "PASS", "stdout must not contain PASS for a malformed config")
	require.NotEmpty(t, stderr.String(), "stderr must contain a parse-failure message")
}

// TestRunMissingArg asserts that omitting the config path exits non-zero and
// prints a usage message to stderr.
func TestRunMissingArg(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"ffparity"}, &stdout, &stderr)

	assert.NotEqual(t, 0, code, "missing argument must exit non-zero")
	assert.Contains(t, stderr.String(), "usage:", "stderr must contain usage instructions")
}
