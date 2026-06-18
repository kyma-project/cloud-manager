package azuremanagedredis

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// modifyKcpAzureManagedRedis patches the KCP AzureManagedRedis spec.sku when the
// SKR redisTier changes within the same family (e.g. S1→S3, P1→P3, C3→C5).
// Cross-family changes are rejected by the API server CEL rule before reaching here.
func modifyKcpAzureManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.KcpAzureManagedRedis == nil {
		return nil, ctx
	}

	skrTier := state.ObjAsAzureManagedRedis().Spec.RedisTier
	desiredSpec, err := TierToSpec(skrTier)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error resolving redisTier to Azure spec", composed.StopWithRequeue, ctx)
	}

	desiredSKU := string(desiredSpec.SKU)
	if state.KcpAzureManagedRedis.Spec.SKU == desiredSKU {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).
		WithValues(
			"currentSKU", state.KcpAzureManagedRedis.Spec.SKU,
			"desiredSKU", desiredSKU,
			"tier", skrTier,
		).
		Info("Patching KCP AzureManagedRedis SKU for tier resize")

	patch := client.MergeFrom(state.KcpAzureManagedRedis.DeepCopy())
	state.KcpAzureManagedRedis.Spec.SKU = desiredSKU
	err = state.KcpCluster.K8sClient().Patch(ctx, state.KcpAzureManagedRedis, patch)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching KCP AzureManagedRedis SKU", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
