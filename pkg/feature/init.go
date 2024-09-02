package feature

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
	"os"
	"time"
)

var provider types.Provider

func init() {
	InitializeFromStaticConfig(nil)
}

type ProviderOptions struct {
	logger    logr.Logger
	loggerSet bool
	filename  string
	url       string
}

type ProviderOption func(o *ProviderOptions)

func WithFile(file string) ProviderOption {
	return func(o *ProviderOptions) {
		o.filename = file
	}
}

func WithUrl(url string) ProviderOption {
	return func(o *ProviderOptions) {
		o.url = url
	}
}

func WithLogger(logger logr.Logger) ProviderOption {
	return func(o *ProviderOptions) {
		o.logger = logger
		o.loggerSet = true
	}
}

func Initialize(ctx context.Context, opts ...ProviderOption) (err error) {
	o := &ProviderOptions{}
	for _, x := range opts {
		x(o)
	}
	if !o.loggerSet {
		o.logger = logr.Discard()
	}
	if len(o.filename) == 0 && len(o.url) == 0 {
		o.url = os.Getenv("FEATURE_FLAG_CONFIG_URL")
		if len(o.url) == 0 {
			o.filename = os.Getenv("FEATURE_FLAG_CONFIG_FILE")
		}
		if len(o.filename) == 0 {
			o.url = "https://raw.githubusercontent.com/kyma-project/cloud-manager/main/config/featureToggles/featureToggles.yaml"
		}
	}
	var rtvr retriever.Retriever
	if o.url != "" {
		o.logger.WithValues("url", o.url).Info("Using http retriever")
		rtvr = &httpretriever.Retriever{URL: o.url}
	} else {
		o.logger.WithValues("file", o.filename).Info("Using file retriever")
		rtvr = &fileretriever.Retriever{Path: o.filename}
	}
	ff, errLoading := ffclient.New(ffclient.Config{
		PollingInterval:         10 * time.Second,
		Context:                 ctx,
		FileFormat:              "yaml",
		EnablePollingJitter:     true,
		StartWithRetrieverError: true,
		Retriever:               rtvr,
	})
	if errLoading != nil {
		err = fmt.Errorf("loading error: %w", errLoading)
	}

	go func() {
		<-ctx.Done()
		ff.Close()
	}()

	provider = &providerGoFF{ff: ff}

	return
}

func InitializeFromStaticConfig(env abstractions.Environment) {
	if env == nil {
		env = abstractions.NewMockedEnvironment(map[string]string{})
	}
	provider = NewProviderConfig(env)
}
