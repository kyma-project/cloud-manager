package awsrediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// kcpRedisClusterProvisioned reports whether the KCP RedisCluster has published its
// connection data (DiscoveryEndpoint). That data — not the transient Ready condition,
// which KCP clears during a modify/auto-heal — is what the auth secret is built from,
// and once published it persists across condition churn, so the secret stays reconciled
// even while KCP is Updating.
func kcpRedisClusterProvisioned(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.KcpRedisCluster != nil &&
		len(state.KcpRedisCluster.Status.DiscoveryEndpoint) > 0
}
