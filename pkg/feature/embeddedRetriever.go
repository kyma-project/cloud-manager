package feature

import (
	"context"
	_ "embed"
)

//go:embed default.yaml
var embeddedConfig []byte

type EmbeddedRetriever struct {
}

func (r *EmbeddedRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	return embeddedConfig, nil
}
