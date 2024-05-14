package nfsinstance

import (
	"context"
	"fmt"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/pointer"
	"time"
)

func loadMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.efs == nil {
		return nil, nil
	}

	mtList, err := state.awsClient.DescribeMountTargets(ctx, pointer.StringDeref(state.efs.FileSystemId, ""))
	if awsmeta.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading mount targets", ctx)
	}

	state.mountTargets = mtList

	state.mountTargetSecurityGroups = make(map[string][]string, len(mtList))
	for _, mt := range mtList {
		// in order not to exhaust rate limit have to slow down
		// in this place usually we get ThrottlingException
		time.Sleep(util.Timing.T10000ms())
		mtID := pointer.StringDeref(mt.MountTargetId, "")
		sgList, err := state.awsClient.DescribeMountTargetSecurityGroups(ctx, mtID)
		if awsmeta.IsNotFound(err) {
			state.mountTargetSecurityGroups[mtID] = []string{}
			continue
		}
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error loading mount target security groups", ctx)
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
