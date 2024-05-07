package iprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func preventCidrChange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	// status.cidr is empty OR same as spec.cidr
	if len(state.ObjAsIpRange().Status.Cidr) == 0 ||
		state.ObjAsIpRange().Spec.Cidr == state.ObjAsIpRange().Status.Cidr {
		return nil, nil
	}

	// status.cidr is not empty AND different from spec.cidr
	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.StateError

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "False",
			Reason:  cloudresourcesv1beta1.ConditionReasonCidrCanNotBeChanged,
			Message: "IpRange CIDR can not be changed",
		}).
		ErrorLogMessage("Error updating IpRange status with CIDR changed condition").
		SuccessLogMsg("Forgetting IpRange with changed Cidr").
		Run(ctx, state)
}
