package azuremanagedredis

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// modifyKcpAzureManagedRedis is intentionally a no-op.
//
// The KCP AzureManagedRedis spec is fully immutable (every field carries a CEL
// `self == oldSelf` rule), and the SKR-side Tier is also immutable. There is
// nothing to mutate after initial creation, so this action exists only for
// symmetry with the AzureRedisInstance / AzureRedisCluster reconcilers, where
// the comparable action handles capacity/SKU changes.
//
// If AMR ever grows mutable fields (e.g. tier upgrade), the resize logic
// belongs here.
func modifyKcpAzureManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	return nil, ctx
}
