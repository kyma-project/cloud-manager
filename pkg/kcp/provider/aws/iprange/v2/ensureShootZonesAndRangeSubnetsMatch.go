package v2

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ensureShootZonesAndRangeSubnetsMatch(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	rangeSubnetCount := len(state.ObjAsIpRange().Status.Ranges)
	shootZonesCount := len(state.Scope().Spec.Scope.Aws.Network.Zones)
	if rangeSubnetCount != shootZonesCount {
		logger = logger.WithValues(
			"rangeSubnetCount", rangeSubnetCount,
			"shootZonesCount", shootZonesCount,
		)

		state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.ErrorState
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudresourcesv1beta1.ReasonShootAndVpcMismatch,
				Message: fmt.Sprintf("RangeSubnetCount %d different then shootZonesCount %d", rangeSubnetCount, shootZonesCount),
			}).
			ErrorLogMessage("Error patching KCP IpRange status on shoot and vpc mismatch").
			SuccessLogMsg("Forgetting KCP IpRange with different subnets and shoot zones count").
			Run(ctx, st)
	}

	return nil, nil
}
