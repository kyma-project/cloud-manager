package v1

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func validateAlways(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If the instance already exists or if it is deleting, continue to next action.
	if state.fsInstance != nil || composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance", nfsInstance.Name).Info("Validating Instance Details")

	//Get GCP details.
	gcpOptions := nfsInstance.Spec.Instance.Gcp

	//Validate whether the requested capacity is a valid value.
	if _, err := IsValidCapacity(gcpOptions.Tier, gcpOptions.CapacityGb); err != nil {
		state.validations = append(state.validations, err.Error())
	}

	if len(state.validations) > 0 {
		return composed.UpdateStatus(nfsInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ReasonValidationFailed,
				Message: strings.Join(state.validations, "\n"),
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("Error validating the Filestore Instance.").
			Run(ctx, state)
	}

	return nil, nil
}
