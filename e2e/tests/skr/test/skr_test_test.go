package gcp

import (
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/cloud-manager/e2e"
)

var opts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Concurrency: 10,
	FS:          e2e.Features,
	Tags:        "@skr && @test",
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestFeatures(t *testing.T) {
	e2e.SkipE2eTests(t)
	o := opts
	o.TestingT = t

	status := godog.TestSuite{
		Name:                 "skr-test",
		Options:              &o,
		TestSuiteInitializer: e2e.InitializeTestSuite,
		ScenarioInitializer:  e2e.InitializeScenario,
	}.Run()

	if status == 2 {
		// command line usage error
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}
