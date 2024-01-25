package nfsinstance

import (
	"context"
	"errors"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func validatePostCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	nfsInstance := state.ObjAsNfsInstance()
	logger.WithValues("NfsInstance :", nfsInstance.Name).Info("Validating Instance Details")

	//Get GCP details.
	gcpOptions := nfsInstance.Spec.Instance.Gcp
	name := nfsInstance.Spec.RemoteRef.Name

	id := nfsInstance.Status.Id
	if id != "" {
		matches := client.FilestoreInstanceRegEx.FindStringSubmatch(id)
		l := len(matches)

		//Validate the location of the instance is not modified
		if l > 2 && matches[2] != gcpOptions.Location {
			state.validations = append(state.validations, "Location cannot be modified")
		}

		//Validate if the name of the instance is not modified.
		if l > 3 && matches[3] != name {
			state.validations = append(state.validations, "Name cannot be modified")
		}
	}

	//Validate the Tier is not modified.
	if state.fsInstance != nil && v1beta1.GcpFileTier(state.fsInstance.Tier) != gcpOptions.Tier {
		state.validations = append(state.validations, "Tier cannot be modified")
	}

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
