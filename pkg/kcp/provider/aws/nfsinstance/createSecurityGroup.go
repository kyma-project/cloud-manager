package nfsinstance

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func createSecurityGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.securityGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	sgId, err := state.awsClient.CreateSecurityGroup(ctx, state.IpRange().Status.VpcId, state.Obj().GetName(), []ec2types.Tag{
		{
			Key:   ptr.To("Name"),
			Value: ptr.To(state.Obj().GetName()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: ptr.To(state.ObjAsNfsInstance().Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: ptr.To(state.Name().String()),
		},
		{
			Key:   ptr.To(common.TagScope),
			Value: ptr.To(state.ObjAsNfsInstance().Spec.Scope.Name),
		},
	})
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error creating security group", ctx)
	}

	state.securityGroupId = sgId

	logger = logger.WithValues("securityGroupId", sgId)
	logger.Info("Security group created")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
