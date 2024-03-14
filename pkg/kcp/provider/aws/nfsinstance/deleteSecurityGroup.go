package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
	"time"
)

func deleteSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.securityGroup == nil {
		return nil, nil
	}

	logger.
		WithValues("securityGroupId", pointer.StringDeref(state.securityGroup.GroupId, "")).
		Info("Deleting security group")

	err := state.awsClient.DeleteSecurityGroup(ctx, pointer.StringDeref(state.securityGroup.GroupId, ""))
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting security group", composed.StopWithRequeueDelay(200*time.Millisecond), ctx)
	}

	return nil, nil
}
