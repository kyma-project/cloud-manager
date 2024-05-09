package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/pointer"
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
		return awsmeta.LogErrorAndReturn(err, "Error deleting security group", ctx)
	}

	return nil, nil
}
