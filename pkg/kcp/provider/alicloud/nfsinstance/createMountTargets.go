package nfsinstance

import (
	"context"

	alicloudnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/nfsinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// createMountTargets ensures a NAS mount target exists in the IpRange VPC.
// AliCloud NAS exposes a single VPC-wide mount target domain, so one mount target
// (bound to the first IpRange vSwitch) is sufficient for the whole VPC.
func createMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.mountTargets) > 0 {
		return nil, ctx
	}

	vpcId := state.IpRange().Status.VpcId
	subnets := state.IpRange().Status.Subnets
	if vpcId == "" || len(subnets) == 0 {
		// validateIpRangeSubnets guards this; defensively requeue.
		return composed.StopWithRequeue, ctx
	}
	vSwitchId := subnets[0].Id

	logger.Info("Creating AliCloud NAS mount target", "fileSystemId", state.fileSystemId, "vpcId", vpcId, "vSwitchId", vSwitchId)

	domain, err := state.client.CreateMountTarget(ctx, state.fileSystemId, vpcId, vSwitchId, state.accessGroupName)
	if err != nil {
		logger.Error(err, "Error creating AliCloud NAS mount target")
		return composed.StopWithRequeue, ctx
	}

	state.mountTargets = append(state.mountTargets, alicloudnfsinstanceclient.MountTargetInfo{
		MountTargetDomain: domain,
		NetworkType:       "Vpc",
		VpcId:             vpcId,
		VSwitchId:         vSwitchId,
		AccessGroup:       state.accessGroupName,
		Status:            "Pending",
	})

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
