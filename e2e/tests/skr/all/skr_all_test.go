package all

import (
	"flag"
	"testing"

	"github.com/cucumber/godog"
	"github.com/kyma-project/cloud-manager/e2e/tests"
)

var opts *godog.Options

func init() {
	opts = tests.BuildOptions()
	godog.BindFlags("godog.", flag.CommandLine, opts)
}

func TestFeatures(t *testing.T) {
	tests.CommonTest(t, opts, "skr-all")
}
