package redisinstance

import (
	"context"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyInstanceClass issues a ModifyInstanceSpec if the desired InstanceClass
// or ReadOnlyCount drift from the observed state. Per issue #2012 design
// decision 8, InstanceClass changes must not be combined with ShardCount
// changes in the same call - this action never touches ShardCount, that is
// handled by the cluster-specific modifyShardCount action.
func modifyInstanceClass(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisInstance()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desiredClass := kcp.Spec.Instance.Alicloud.InstanceClass
	desiredReadOnly := kcp.Spec.Instance.Alicloud.ReadOnlyCount

	classDrift := desiredClass != "" && desiredClass != state.instance.InstanceClass
	// Some classes (tair.rdb.*, redis.amber.master.*.multithread) silently ignore
	// ReadOnlyCount - the field is absent from DescribeInstanceAttribute and
	// ModifyInstanceSpec has no effect. Skip drift detection for those classes to
	// avoid an infinite modify loop.
	readOnlyDrift := !alicloudclient.IsReadOnlyCountUnsupported(state.instance.InstanceClass) &&
		desiredReadOnly != state.instance.ReadOnlyCount
	if !classDrift && !readOnlyDrift {
		return nil, ctx
	}

	opts := alicloudclient.ModifyInstanceSpecOptions{}
	if classDrift {
		opts.InstanceClass = desiredClass
	}
	// Only include ReadOnlyCount when it has actually drifted. Some instance
	// classes (e.g. tair.rdb.*) reject any ReadOnlyCount value via API
	// (COMMODITY.INVALID_COMPONENT), so sending it unconditionally would break
	// every class-only modification on those classes.
	if readOnlyDrift {
		opts.ReadOnlyCount = tea.Int32(desiredReadOnly)
	}

	if err := state.client.ModifyInstanceSpec(ctx, state.instance.InstanceId, opts); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore instance spec",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	// Force re-load next reconcile.
	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
