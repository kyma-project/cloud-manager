package feature

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"os"
	"time"
)

var provider types.Provider

func init() {
	InitializeFromStaticConfig(nil)
}

type ProviderOptions struct {
	filename string
}

type ProviderOption func(o *ProviderOptions)

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
		o.filename = os.Getenv("FEATURE_FLAG_CONFIG_FILE")
		if len(o.filename) == 0 {
			o.filename = "/var/cloud-manager.kyma-project.io/config/featureFlags.yaml"
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

	provider = &providerGoFF{ff: ff}

	return
}

func InitializeFromStaticConfig(env abstractions.Environment) {
	if env == nil {
		env = abstractions.NewMockedEnvironment(map[string]string{})
	}
	provider = NewProviderConfig(env)
}
