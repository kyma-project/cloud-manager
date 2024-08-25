package nfsinstance

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/shares"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func shareNetworkDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.shareNetwork == nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	arr, err := state.cceeClient.ListShares(ctx, state.shareNetwork.ID)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing shares for delete share network", composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}
	if len(arr) > 1 {
		ids := pie.Map(arr, func(sh shares.Share) string {
			return fmt.Sprintf("%s-%s", sh.ID, sh.Name)
		})
		logger.
			WithValues("existingShares", fmt.Sprintf("%v", ids)).
			Info("Other shares exist, not deleting CCEE share network")

		return nil, nil
	}

	logger.Info("Deleting CCEE share")

	err = state.cceeClient.DeleteShareNetwork(ctx, state.shareNetwork.ID)

	if err != nil && !gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ErrorState
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error deleting CCEE shareNetwork",
			}).
			FailedError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			ErrorLogMessage("Error patching CCEE NfsInstance status after error deleting shareNetwork").
			Run(ctx, state)
	}

	state.ObjAsNfsInstance().SetStateData(StateDataShareNetworkId, "")

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error patching CCEE NfsInstance status after shareNetwork delete").
		SuccessErrorNil().
		Run(ctx, state)
}
