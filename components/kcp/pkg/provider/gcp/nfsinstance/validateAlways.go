package nfsinstance

import (
	"context"
	"errors"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func validateAlways(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	//If the instance already exists, continue.
	if state.fsInstance != nil {
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

	//Validate whether the nwMask is a valid value.
	cidr := ""
	if state.IpRange() != nil {
		cidr = state.IpRange().Spec.Cidr
	}
	if _, err := IsValidNwMask(gcpOptions.Tier, cidr); err != nil {
		state.validations = append(state.validations, err.Error())
	}

	if len(state.validations) > 0 {
		err := errors.New(strings.Join(state.validations, "\n"))
		state.AddErrorCondition(ctx, v1beta1.ReasonValidationFailed, err)
		return composed.LogErrorAndReturn(err, "Validation Failed", composed.StopAndForget, nil)
	}

	return nil, nil
}
