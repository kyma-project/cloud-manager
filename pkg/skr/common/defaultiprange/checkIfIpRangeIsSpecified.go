package defaultiprange

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkIfIpRangeIsSpecified(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	if feature.IpRangeAutomaticCidrAllocation.Value(ctx) {
		return nil, nil
	}

	if len(state.ObjAsObjWithIpRangeRef().GetIpRangeRef().Name) > 0 {
		return nil, nil
	}

	state.ObjAsObjWithIpRangeRef().SetState(cloudresourcesv1beta1.StateError)
	var sb *composed.UpdateStatusBuilder
	if _, ok := state.ObjAsObjWithIpRangeRef().(composed.ObjWithCloneForPatchStatus); ok {
		sb = composed.PatchStatus(state.ObjAsObjWithIpRangeRef())
	} else {
		sb = composed.UpdateStatus(state.ObjAsObjWithIpRangeRef())
	}
	return sb.
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeNotFound,
			Message: "IpRangeRef is required",
		}).
		SuccessLogMsg(fmt.Sprintf("Forgetting SKR %T with empty IpRangeRef when IpRangeAutomaticCidrAllocation is disabled", state.ObjAsObjWithIpRangeRef())).
		ErrorLogMessage(fmt.Sprintf("Error patching SKR %T status with error for empty IpRangeRef when IpRangeAutomaticCidrAllocation is disabled", state.ObjAsObjWithIpRangeRef())).
		Run(ctx, state)
}
