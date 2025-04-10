package nfsinstance

import (
	"context"
	"fmt"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

// removeMountTargetsFromOtherVpcs is a migration fix for a bug when wrong VPC was choosen to create EFS in.
// If there are more then 3 mount targets, it will load subnet for all of them and delete those from EFS
// that belong to a VPC other than the one specified in the ipRange.status.vpcId
func removeMountTargetsFromOtherVpcs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.mountTargets) <= 3 {
		return nil, nil
	}

	logger.
		WithValues(
			"mountTargets",
			fmt.Sprintf("%v", pie.Map(state.mountTargets, func(mt efstypes.MountTargetDescription) string {
				mtID := ptr.Deref(mt.MountTargetId, "")
				return fmt.Sprintf(
					"{id:%s, az:%s, ip: %s, sg: %v}",
					mtID,
					ptr.Deref(mt.AvailabilityZoneId, ""),
					ptr.Deref(mt.IpAddress, ""),
					state.mountTargetSecurityGroups[mtID],
				)
			})),
		).
		Info("Exploring mount targets to remove")

	var mountTargetsToRemove []string
	for _, mt := range state.mountTargets {
		subnetId := ptr.Deref(mt.SubnetId, "")
		subnet, err := state.awsClient.DescribeSubnet(ctx, subnetId)
		if err != nil {
			if awsmeta.IsErrorRetryable(err) {
				return awsmeta.LogErrorAndReturn(err, "Retryable error describing subnet", ctx)
			}
			logger.
				WithValues("subnetId", subnetId).
				Error(err, "Error describing subnet")
			continue
		}
		if subnet == nil {
			logger.
				WithValues("subnetId", subnetId).
				Info("EFS mount target subnet does not exist")
			continue
		}

		vpcId := ptr.Deref(subnet.VpcId, "")
		if vpcId != "" && vpcId != state.IpRange().Status.VpcId {
			mountTargetsToRemove = append(mountTargetsToRemove, ptr.Deref(mt.MountTargetId, ""))
		}
	}

	logger.
		WithValues(
			"mountTargetsToRemove",
			fmt.Sprintf("%v", mountTargetsToRemove),
		).
		Info("Mount Targets from other VPC to remove")

	for _, mtId := range mountTargetsToRemove {
		logger.
			WithValues("mountTargetId", mtId).
			Info("Removing mount target from other VPC")
		err := state.awsClient.DeleteMountTarget(ctx, mtId)
		if err != nil {
			if awsmeta.IsErrorRetryable(err) {
				return awsmeta.LogErrorAndReturn(err, "Retryable error removing mount target from other VPC", ctx)
			}
			logger.
				WithValues("mountTargetId", mtId).
				Error(err, "Error removing mount target from other VPC")
		}
	}

	return nil, nil
}
