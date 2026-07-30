package rediscluster

import (
	"context"
	"regexp"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// proxyShardTokensRe matches the shard-count and proxy-count tokens embedded in
// proxy class names, e.g. ".4db.0rodb.8proxy." in
// "redis.logic.sharding.4g.4db.0rodb.8proxy.default".
var proxyShardTokensRe = regexp.MustCompile(`\.\d+db\.0rodb\.\d+proxy\.`)

// proxyClassTierKey returns a canonical key for a proxy class that omits the
// shard-count and proxy-count tokens (both vary together with shardCount), so
// that class names that differ only because shardCount changed are treated as
// the same tier. Returns the original string for non-proxy classes.
func proxyClassTierKey(class string) string {
	if !alicloudclient.IsProxyClusterClass(class) {
		return class
	}
	return proxyShardTokensRe.ReplaceAllLiteralString(class, ".<N>db.0rodb.<N>proxy.")
}

// modifyInstanceClass issues ModifyInstanceSpec if the desired InstanceClass
// drifts from the observed state in a way that requires an API call.
//
// For proxy-based classes (redis.logic.sharding.*) the class name encodes the
// shard count. When a user changes only shardCount, redisTierToInstanceClass
// produces a new class name with the new shard count embedded. This creates a
// spurious class-name drift that must NOT trigger ModifyInstanceSpec — the
// actual shard count change is handled exclusively by modifyShardCount. Only
// tier-level drift (different memory size) warrants a ModifyInstanceSpec call.
//
// The replica-drift guard uses the observed class (state.instance.InstanceClass)
// rather than the desired class because the instance may still be running on the
// old (proxy) class during an in-flight tier change.
func modifyInstanceClass(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desiredClass := kcp.Spec.Instance.Alicloud.InstanceClass
	observedClass := state.instance.InstanceClass
	desiredReplicas := kcp.Spec.Instance.Alicloud.ReplicasPerShard

	// For proxy classes, compare tier keys (strip the shard-count token).
	// A difference in the embedded shard count alone does not need ModifyInstanceSpec;
	// modifyShardCount handles it via AddShardingNode/DeleteShardingNode.
	classDrift := desiredClass != "" && proxyClassTierKey(desiredClass) != proxyClassTierKey(observedClass)

	// Use the observed class to decide whether replicas are tunable: the instance
	// may still be on its old (proxy) class if a tier change is in progress.
	replicasDrift := !alicloudclient.IsProxyClusterClass(observedClass) &&
		desiredReplicas != state.instance.ReadOnlyCount
	if !classDrift && !replicasDrift {
		return nil, ctx
	}

	opts := alicloudclient.ModifyInstanceSpecOptions{}
	if classDrift {
		opts.InstanceClass = desiredClass
	}
	if replicasDrift {
		opts.ReadOnlyCount = &desiredReplicas
	}

	if err := state.client.ModifyInstanceSpec(ctx, state.instance.InstanceId, opts); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster instance class",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
