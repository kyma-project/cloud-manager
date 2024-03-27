package nfsinstance

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func createSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	sgId, err := state.awsClient.CreateSecurityGroup(ctx, state.IpRange().Status.VpcId, state.Obj().GetName(), []ec2Types.Tag{
		{
			Key:   pointer.String("Name"),
			Value: pointer.String(state.Obj().GetName()),
		},
		{
			Key:   pointer.String(common.TagCloudManagerRemoteName),
			Value: pointer.String(state.ObjAsNfsInstance().Spec.RemoteRef.String()),
		},
		{
			Key:   pointer.String(common.TagCloudManagerName),
			Value: pointer.String(state.Name().String()),
		},
		{
			Key:   pointer.String(common.TagScope),
			Value: pointer.String(state.ObjAsNfsInstance().Spec.Scope.Name),
		},
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating security group", composed.StopWithRequeue, ctx)
	}

	state.securityGroupId = sgId

	logger = logger.WithValues("securityGroupId", sgId)
	logger.Info("Security group created")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
