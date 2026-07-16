package nfsinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateIpRangeSubnets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.IpRange() != nil && len(state.IpRange().Status.Subnets) > 0 && state.IpRange().Status.VpcId != "" {
		return nil, ctx
	}

	return composed.UpdateStatus(state.ObjAsNfsInstance()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonValidationFailed,
			Message: "IpRange has no subnets or VPC, and at least one subnet is needed",
		}).
		SuccessError(composed.StopAndForget).
		Run(ctx, st)
}
