package clean

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithDryRun(dryRun bool) Option {
	return func(o *options) {
		o.dryRun = dryRun
	}
}

func WithClient(client client.Client) Option {
	return func(o *options) {
		o.client = client
	}
}

func WithScheme(scheme *runtime.Scheme) Option {
	return func(o *options) {
		o.scheme = scheme
	}
}

func WithMatchers(matchers ...Matcher) Option {
	return func(o *options) {
		o.matchers = append(o.matchers, matchers...)
	}
}

func WithWait(wait bool) Option {
	return func(o *options) {
		o.wait = wait
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

func WithPollInterval(pollInterval time.Duration) Option {
	return func(o *options) {
		o.pollInterval = pollInterval
	}
}

func WithSleeper(sleeper util.Sleeper) Option {
	return func(o *options) {
		o.sleeper = sleeper
	}
}

func WithForceDeleteOnTimeout(v bool) Option {
	return func(o *options) {
		o.forceDeleteOnTimeout = v
	}
}

func WithLogger(logger logr.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}
