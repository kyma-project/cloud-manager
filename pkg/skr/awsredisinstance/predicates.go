package awsredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// kcpRedisInstanceProvisioned reports whether the KCP RedisInstance has published its
// connection data (PrimaryEndpoint). That data — not the transient Ready condition,
// which KCP clears during a modify/auto-heal — is what the auth secret is built from,
// and once published it persists across condition churn, so the secret stays reconciled
// even while KCP is Updating.
func kcpRedisInstanceProvisioned(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.KcpRedisInstance != nil &&
		len(state.KcpRedisInstance.Status.PrimaryEndpoint) > 0
}
