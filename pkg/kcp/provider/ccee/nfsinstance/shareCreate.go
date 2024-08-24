package nfsinstance

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shareCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.share != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Creating CCEE share")

	metadata := map[string]string{
		common.TagCloudManagerName:       state.Obj().GetName(),
		common.TagCloudManagerRemoteName: state.ObjAsNfsInstance().Spec.RemoteRef.Name,
		common.TagScope:                  state.ObjAsNfsInstance().Spec.Scope.Name,
		common.TagShoot:                  state.Scope().Spec.ShootName,
	}
	share, err := state.cceeClient.CreateShare(
		ctx,
		state.shareNetwork.ID,
		state.ShareName(),
		state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb,
		"",
		metadata,
	)
	if err != nil {
		logger.Error(err, "Error creating CCEE share")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ErrorState
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error creating CCEE share",
			}).
			ErrorLogMessage("Error patching CCEE NfsInstance status with error state after failed share creation").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.share = share

	state.ObjAsNfsInstance().SetStateData(StateDataShareId, share.ID)
	state.ObjAsNfsInstance().Status.Id = share.ID

	state.ObjAsNfsInstance().Status.State = "Creating"
	state.ObjAsNfsInstance().Status.CapacityGb = state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error updating CCEE NfsInstance state data with created shareId").
		SuccessErrorNil().
		Run(ctx, state)
}
