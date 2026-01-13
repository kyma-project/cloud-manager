package tests

import (
	"flag"
	"testing"

	"github.com/cucumber/godog"
)

var opts *godog.Options

func init() {
	// @none tag should not exist on any scenario, so this will not run any test unless cli flag provided
	opts = BuildOptions(WithTags("@none && ~@skip"))
	godog.BindFlags("godog.", flag.CommandLine, opts)
}

func TestFeatures(t *testing.T) {
	CommonTest(t, opts, "features")
}
