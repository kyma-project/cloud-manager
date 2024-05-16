package feature

import (
	"context"
	"fmt"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"os"
	"time"
)

var provider *ffclient.GoFeatureFlag

type ProviderOptions struct {
	filename              string
	evaluateAllFlagsState bool
}

type ProviderOption func(o *ProviderOptions)

func WithEvaluateAllFlagsState() ProviderOption {
	return func(o *ProviderOptions) {
		o.evaluateAllFlagsState = true
	}
}

func WithFile(file string) ProviderOption {
	return func(o *ProviderOptions) {
		o.filename = file
	}
}

func Initialize(ctx context.Context, opts ...ProviderOption) (err error) {
	o := &ProviderOptions{}
	for _, x := range opts {
		x(o)
	}
	if len(o.filename) == 0 {
		o.filename = os.Getenv("FEATURE_FLAG_FILE_CONFIG")
		if len(o.filename) == 0 {
			o.filename = "/var/cloud-manager.kyma-project.io/featureFlags.yaml"
		}
	}
	ff, errLoading := ffclient.New(ffclient.Config{
		PollingInterval:         10 * time.Second,
		Context:                 ctx,
		FileFormat:              "yaml",
		EnablePollingJitter:     true,
		StartWithRetrieverError: true,
		Retriever:               &fileretriever.Retriever{Path: o.filename},
	})
	if errLoading != nil {
		err = fmt.Errorf("loading error: %w", errLoading)
	}

	go func() {
		<-ctx.Done()
		ff.Close()
	}()

	if o.evaluateAllFlagsState {
		allFlags := ff.AllFlagsState(ffcontext.NewEvaluationContext(""))
		var failures []string
		for name, flag := range allFlags.GetFlags() {
			if flag.Failed {
				failures = append(failures, fmt.Sprintf("%s: %s", name, flag.ErrorCode))
			}
		}
		if len(failures) > 0 {
			errFlag := fmt.Errorf("failed flags: %v", failures)
			if err == nil {
				err = errFlag
			} else {
				err = fmt.Errorf("errors: %w, %w", err, errFlag)
			}
		}
	}
	provider = ff

	return
}
