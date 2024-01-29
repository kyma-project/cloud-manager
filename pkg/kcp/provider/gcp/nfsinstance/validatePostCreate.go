package nfsinstance

import (
	"context"
	"errors"
	"strings"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func validatePostCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If it is deleting, continue to next action
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Validating Instance Details")

	//Get GCP details.
	gcpOptions := nfsInstance.Spec.Instance.Gcp

	//Validate the instance is not being scale down.
	if !CanScaleDown(gcpOptions.Tier) && state.fsInstance != nil &&
		state.fsInstance.FileShares[0].CapacityGb > int64(gcpOptions.CapacityGb) {
		state.validations = append(state.validations, "Capacity cannot be reduced.")
	}

	//Add error condition
	if len(state.validations) > 0 {
		err := errors.New(strings.Join(state.validations, "\n"))
		state.AddErrorCondition(ctx, v1beta1.ReasonValidationFailed, err)
		return composed.LogErrorAndReturn(err, "Validation Failed", composed.StopAndForget, nil)
	}

	return nil, nil
}
