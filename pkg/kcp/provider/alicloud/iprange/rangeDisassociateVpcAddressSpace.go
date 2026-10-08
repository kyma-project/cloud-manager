package iprange

import (
	"context"
	"slices"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	ipclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func rangeDisassociateVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	cidr := state.ObjAsIpRange().Status.Cidr
	if cidr == "" {
		return nil, ctx
	}

	if !slices.Contains(state.secondaryCidrBlocks, cidr) {
		return nil, ctx
	}

	logger.Info("Disassociating secondary CIDR block from AliCloud VPC", "vpcId", state.vpcId, "cidr", cidr)

	err := state.client.UnassociateVpcCidrBlock(ctx, state.vpcId, cidr)
	if err != nil {
		if ipclient.IsCidrInUseErr(err) {
			logger.Error(err, "CIDR block still in use by a vSwitch - manual vSwitch deletion required before IpRange can be deleted",
				"vpcId", state.vpcId, "cidr", cidr)

			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:               cloudcontrolv1beta1.ConditionTypeError,
					Status:             metav1.ConditionTrue,
					ObservedGeneration: state.ObjAsIpRange().Generation,
					Reason:             cloudcontrolv1beta1.ReasonFailedDisassociatingVpcAddressSpace,
					Message:            "CIDR block is still in use by a vSwitch; delete the vSwitch in the cloud console to unblock deletion",
				}).
				ErrorLogMessage("Error patching AliCloud KCP IpRange status after CidrInUse error").
				FailedError(composed.StopWithRequeue).
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
				Run(ctx, state)
		}
		logger.Error(err, "Error disassociating secondary CIDR block from AliCloud VPC")
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
