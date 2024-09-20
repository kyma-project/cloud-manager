package iprange

import (
	"context"
	"fmt"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func securityGroupCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.securityGroup != nil {
		return nil, ctx
	}

	logger.Info("Creating Azure KCP IpRange security group")

	err := state.azureClient.CreateSecurityGroup(
		ctx, state.resourceGroupName, state.securityGroupName,
		state.Scope().Spec.Region,
		nil,
		map[string]string{
			common.TagScope:                  fmt.Sprintf("%s/%s", state.Scope().Namespace, state.Scope().Name),
			common.TagShoot:                  state.Scope().Spec.ShootName,
			common.TagCloudManagerRemoteName: state.ObjAsIpRange().Spec.RemoteRef.String(),
			common.TagCloudManagerName:       fmt.Sprintf("%s/%s", state.ObjAsIpRange().Namespace, state.ObjAsIpRange().Name),
		},
	)

	if azuremeta.IsTooManyRequests(err) {
		return azuremeta.LogErrorAndReturn(err, "Azure KCP IpRange too many requests on create security group", ctx)
	}

	if err != nil {
		state.ObjAsIpRange().Status.State = cloudcontrol1beta1.ErrorState

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrol1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrol1beta1.ConditionTypeError,
				Message: "Error creating Azure security group",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange status after failed creating security group").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
