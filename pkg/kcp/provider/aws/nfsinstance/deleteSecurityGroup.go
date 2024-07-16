package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.securityGroup == nil {
		return nil, nil
	}

	logger.
		WithValues("securityGroupId", ptr.Deref(state.securityGroup.GroupId, "")).
		Info("Deleting security group")

	err := state.awsClient.DeleteSecurityGroup(ctx, ptr.Deref(state.securityGroup.GroupId, ""))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting security group", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
