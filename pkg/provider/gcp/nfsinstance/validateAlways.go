package nfsinstance

import (
	"context"
	"errors"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
)

func validateAlways(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If the instance already exists or if it is deleting, continue to next action.
	if state.fsInstance != nil || composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Validating Instance Details")

	//Get GCP details.
	gcpOptions := nfsInstance.Spec.Instance.Gcp

	//Validate whether the requested capacity is a valid value.
	if _, err := IsValidCapacity(gcpOptions.Tier, gcpOptions.CapacityGb); err != nil {
		state.validations = append(state.validations, err.Error())
	}

	if len(state.validations) > 0 {
		err := errors.New(strings.Join(state.validations, "\n"))
		state.AddErrorCondition(ctx, v1beta1.ReasonValidationFailed, err)
		return composed.LogErrorAndReturn(err, "Validation Failed", composed.StopAndForget, nil)
	}

	return nil, nil
}
