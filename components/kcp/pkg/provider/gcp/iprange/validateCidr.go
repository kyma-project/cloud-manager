package iprange

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateCidr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := state.ObjAsIpRange()
	logger.WithValues("ipRange :", ipRange.Name).Info("Loading GCP Address")

	//Parse CIDR.
	addr, prefix, err := util.CidrParseIPnPrefix(ipRange.Spec.Cidr)
	if err != nil {
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonInvalidCidr,
			Message: fmt.Sprint(err),
		})
		state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.ErrorState
		err := state.UpdateObjStatus(ctx)
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
