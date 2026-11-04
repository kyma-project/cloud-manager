package nfsinstance

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func loadMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.efs == nil {
		return nil, ctx
	}

	mtList, err := state.awsClient.DescribeMountTargets(ctx, ptr.Deref(state.efs.FileSystemId, ""))
	if awsmeta.IsNotFound(err) {
		return nil, ctx
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
		mtID := ptr.Deref(mt.MountTargetId, "")
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

	return nil, ctx
}
