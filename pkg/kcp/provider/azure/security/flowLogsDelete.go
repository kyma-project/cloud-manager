package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func flowLogsDelete(ctx context.Context, st composed.State) (error, context.Context) {
	// Flow logs deferred - requires VNet resource ID source to be determined
	return nil, ctx
}
