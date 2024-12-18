package feature

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
)

var provider types.Provider

func init() {
	InitializeFromStaticConfig(nil)
}

type ProviderOptions struct {
	filename string
	url      string
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

func Initialize(ctx context.Context, logger logr.Logger, opts ...ProviderOption) (err error) {
	o := &ProviderOptions{}
	for _, x := range opts {
		x(o)
	}

	if len(o.filename) == 0 && len(o.url) == 0 {
		o.url = os.Getenv("FEATURE_FLAG_CONFIG_URL")
		if len(o.url) == 0 {
			o.filename = os.Getenv("FEATURE_FLAG_CONFIG_FILE")
		}
		if len(o.filename) == 0 {
			return errors.New("unable to locate feature flag config file. Use env vars FEATURE_FLAG_CONFIG_URL or FEATURE_FLAG_CONFIG_FILE to specify its url/path")
		}
	}
	var rtvr retriever.Retriever
	if o.url != "" {
		logger.WithValues("ffRetrieverUrl", o.url).Info("Using http retriever")
		rtvr = &httpretriever.Retriever{URL: o.url}
	} else {
		logger.WithValues("ffRetrieverFile", o.filename).Info("Using file retriever")
		rtvr = &fileretriever.Retriever{Path: o.filename}
	}
	ff, errLoading := ffclient.New(ffclient.Config{
		PollingInterval:         30 * time.Second,
		Context:                 ctx,
		FileFormat:              "yaml",
		EnablePollingJitter:     true,
		StartWithRetrieverError: true,
		Retriever:               rtvr,
		LeveledLogger:           slog.New(logr.ToSlogHandler(logger)),
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
