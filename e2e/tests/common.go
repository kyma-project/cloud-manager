package tests

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/cloud-manager/e2e"
)

type GodogOption func(*godog.Options)

func WithTags(tags string) GodogOption {
	return func(o *godog.Options) {
		o.Tags = tags
	}
}

func BuildOptions(opts ...GodogOption) *godog.Options {
	o := &godog.Options{
		Output:      colors.Colored(os.Stdout),
		Concurrency: 10,
		FS:          e2e.Features,
		Tags:        "@skr && @aws && ~@peering",
		Format:      "pretty",
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func CommonTest(t *testing.T, opts *godog.Options, name string) {
	e2e.SkipE2eTests(t)
	o := opts
	o.TestingT = t

	status := godog.TestSuite{
		Name:                 name,
		Options:              o,
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
