package iprange

import (
	"context"
	"errors"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"google.golang.org/api/googleapi"
)

func loadAddress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Loading GCP Address")

	//Get from GCP.
	gcpScope := state.Scope().Spec.Scope.Gcp
	project := gcpScope.Project
	vpc := gcpScope.VpcNetwork

	addr, err := state.computeClient.GetIpRange(ctx, project, ipRange.Name)
	if err != nil {

		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				state.address = nil
				return nil, nil
			}
		}
		state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, err)
		return composed.LogErrorAndReturn(err, "Error getting Addresses from GCP", composed.StopWithRequeue, nil)
	}

	//Check whether the IPRange is in the same VPC as that of the SKR.
	if !strings.HasSuffix(addr.Network, vpc) {
		state.AddErrorCondition(ctx, v1beta1.ReasonGcpError, errors.New("IPRange with the same name exists in another VPC."))
		return composed.LogErrorAndReturn(err, "GCP - IPRange name conflict", composed.StopWithRequeue, nil)
	}
	state.address = addr

	return nil, nil
}
