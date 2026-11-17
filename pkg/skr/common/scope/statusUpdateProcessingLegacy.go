package scope

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func statusUpdateProcessingLegacy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	if state.ObjWithConditionsAndState() == nil {
		// unsupported type, likely newer API style with object
		// implementing composed.ObjWithStatus
		return nil, ctx
	}

	//If the object already have a state, continue...
	if state.ObjWithConditionsAndState().State() != "" {
		return nil, ctx
	}

	state.ObjWithConditionsAndState().SetState(cloudresourcesv1beta1.StateProcessing)

	return composed.UpdateStatus(state.ObjWithConditionsAndState()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeProcessing,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.StateProcessing,
			Message: cloudresourcesv1beta1.StateProcessing,
		}).
		SuccessErrorNil().
		ErrorLogMessage(fmt.Sprintf("Error setting object %T state to Processing", state.Obj())).
		Run(ctx, state)
}
