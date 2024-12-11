package feature

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

const nukeBackupsGcpFlagName = "nukeBackupsGcp"

var FFNukeBackupsGcp = &nukeBackupsGcpInfo{}

type nukeBackupsGcpInfo struct{}

func (k *nukeBackupsGcpInfo) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, nukeBackupsGcpFlagName, false)
}

func (k *nukeBackupsGcpInfo) Predicate() composed.Predicate {
	return func(ctx context.Context, _ composed.State) bool {
		return k.Value(ctx)
	}
}
