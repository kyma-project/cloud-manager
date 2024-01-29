package iprange

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func validateCidr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Loading GCP Address")

	//Parse CIDR.
	addr, prefix, err := util.CidrParseIPnPrefix(ipRange.Spec.Cidr)
	if err != nil {
		err := state.AddErrorCondition(ctx, cloudresourcesv1beta1.ReasonInvalidCidr, err)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating IpRange status due to cidr overlap", composed.StopWithRequeue, nil)
		}

		return composed.StopAndForget, nil
	}

	//Store the parsed values in the state object.
	state.ipAddress = addr
	state.prefix = prefix

	return nil, nil
}
