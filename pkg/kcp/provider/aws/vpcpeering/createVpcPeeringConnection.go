package vpcpeering

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering != nil {
		return nil, nil
	}

	tags := []ec2types.Tag{
		{
			Key:   ptr.To("Name"),
			Value: ptr.To(state.Obj().GetName()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerName),
			Value: ptr.To(state.Name().String()),
		},
		{
			Key:   ptr.To(common.TagCloudManagerRemoteName),
			Value: ptr.To(state.ObjAsVpcPeering().Spec.RemoteRef.String()),
		},
		{
			Key:   ptr.To(common.TagScope),
			Value: ptr.To(state.ObjAsVpcPeering().Spec.Scope.Name),
		},
		{
			Key:   ptr.To(common.TagShoot),
			Value: ptr.To(state.Scope().Spec.ShootName),
		},
	}

	remoteRegion := state.remoteNetwork.Spec.Network.Reference.Aws.Region

	vpcPeering, err := state.client.CreateVpcPeeringConnection(
		ctx,
		state.vpc.VpcId,
		state.remoteVpc.VpcId,
		ptr.To(remoteRegion),
		state.remoteVpc.OwnerId,
		tags)

	if err != nil {
		logger.Error(err, "Error creating AWS VPC Peering")

		if awsmeta.IsErrorRetryable(err) {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}

		changed := false

		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
			changed = true
		}

		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: fmt.Sprintf("Failed creating VpcPeerings. %s", awsmeta.GetErrorMessage(err)),
		}) {
			changed = true
		}

		if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.WarningState) {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.WarningState)
		}

		if !changed {
			return composed.StopWithRequeueDelay(util.Timing.T300000ms()), nil
		}

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error updating VpcPeering status due to failed creating vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	logger = logger.WithValues("id", ptr.Deref(vpcPeering.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("AWS VPC Peering Connection created")

	state.vpcPeering = vpcPeering

	state.ObjAsVpcPeering().Status.Id = ptr.Deref(vpcPeering.VpcPeeringConnectionId, "")

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error updating VPC Peering status with connection id").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
