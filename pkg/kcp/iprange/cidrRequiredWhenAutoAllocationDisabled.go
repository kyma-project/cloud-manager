package iprange

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cidrRequiredWhenAutoAllocationDisabled(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	autoCidrAllocationEnabled := feature.IpRangeAutomaticCidrAllocation.Value(ctx)
	if autoCidrAllocationEnabled {
		return nil, nil
	}

	if len(state.ObjAsIpRange().Spec.Cidr) > 0 {
		// already allocated
		return nil, nil
	}

	return composed.PatchStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonInvalidCidr,
			Message: "CIDR field is required",
		}).
		SuccessLogMsg("Forgetting KCP IpRange with empty cidr when autoCidrAllocation is disabled").
		ErrorLogMessage("Error patching KCP IpRange status with error on  empty cidr when autoCidrAllocation is disabled").
		Run(ctx, state)
}
