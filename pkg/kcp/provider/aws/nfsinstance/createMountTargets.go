package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/pointer"
)

func createMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	mountTargetsBySubnetId := make(map[string]string, len(state.IpRange().Status.Subnets))
	for _, mt := range state.mountTargets {
		mountTargetsBySubnetId[pointer.StringDeref(mt.SubnetId, "")] = pointer.StringDeref(mt.MountTargetId, "")
	}

	for _, subnet := range state.IpRange().Status.Subnets {
		_, ok := mountTargetsBySubnetId[subnet.Id]
		if ok {
			continue
		}

		logger.
			WithValues(
				"subnetId", subnet.Id,
				"subnetZone", subnet.Zone,
			).
			Info("Creating mount target")

		_, err := state.awsClient.CreateMountTarget(
			ctx,
			pointer.StringDeref(state.efs.FileSystemId, ""),
			subnet.Id,
			[]string{state.securityGroupId},
		)
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error creating Mount point", ctx)
		}
	}

	return nil, nil
}
