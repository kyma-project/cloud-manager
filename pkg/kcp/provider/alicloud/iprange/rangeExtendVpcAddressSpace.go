package iprange

import (
	"context"
	"slices"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func rangeExtendVpcAddressSpace(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	cidr := state.ObjAsIpRange().Status.Cidr

	// Already associated
	if slices.Contains(state.secondaryCidrBlocks, cidr) {
		return nil, ctx
	}

	logger.Info("Associating secondary CIDR block to AliCloud VPC", "vpcId", state.vpcId, "cidr", cidr)

	err := state.client.AssociateVpcCidrBlock(ctx, state.vpcId, cidr)
	if err != nil {
		logger.Error(err, "Error associating secondary CIDR block to AliCloud VPC")
		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:               cloudcontrolv1beta1.ConditionTypeError,
				Status:             metav1.ConditionTrue,
				ObservedGeneration: state.ObjAsIpRange().Generation,
				Reason:             cloudcontrolv1beta1.ReasonFailedExtendingVpcAddressSpace,
				Message:            "Failed associating secondary CIDR block to VPC",
			}).
			ErrorLogMessage("Error patching AliCloud KCP IpRange status after failed CIDR block association").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	state.secondaryCidrBlocks = append(state.secondaryCidrBlocks, cidr)

	state.ObjAsIpRange().Status.AddressSpaceId = cidr

	return composed.PatchStatus(state.ObjAsIpRange()).
		ErrorLogMessage("Error patching AliCloud KCP IpRange status after CIDR block association").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
