package nfsinstance

import (
	"context"
	"fmt"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/utils/pointer"
)

func loadMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	mtList, err := state.awsClient.DescribeMountTargets(ctx, pointer.StringDeref(state.efs.FileSystemId, ""))
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading mount targets", composed.StopWithRequeue, nil)
	}

	state.mountTargets = mtList

	state.mountTargetSecurityGroups = make(map[string][]string, len(mtList))
	for _, mt := range mtList {
		mtID := pointer.StringDeref(mt.MountTargetId, "")
		sgList, err := state.awsClient.DescribeMountTargetSecurityGroups(ctx, mtID)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading mount target security groups", composed.StopWithRequeue, nil)
		}
		state.mountTargetSecurityGroups[mtID] = sgList
	}

	logger.
		WithValues(
			"mountTargets",
			fmt.Sprintf("%v", pie.Map(mtList, func(mt efsTypes.MountTargetDescription) string {
				mtID := pointer.StringDeref(mt.MountTargetId, "")
				return fmt.Sprintf(
					"{id:%s, az:%s, ip: %s, sg: %v}",
					mtID,
					pointer.StringDeref(mt.AvailabilityZoneId, ""),
					pointer.StringDeref(mt.IpAddress, ""),
					state.mountTargetSecurityGroups[mtID],
				)
			})),
		).
		Info("Mount targets loaded")

	return nil, nil
}
