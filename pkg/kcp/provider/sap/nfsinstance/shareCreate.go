package nfsinstance

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shareCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.share != nil {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Creating SAP share")

	metadata := map[string]string{
		common.TagCloudManagerName:       state.Obj().GetName(),
		common.TagCloudManagerRemoteName: state.ObjAsNfsInstance().Spec.RemoteRef.Name,
		common.TagScope:                  state.ObjAsNfsInstance().Spec.Scope.Name,
		common.TagShoot:                  state.Scope().Spec.ShootName,
	}
	share, err := state.sapClient.CreateShareOp(
		ctx,
		state.shareNetwork.ID,
		state.ShareName(),
		state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb,
		"",
		metadata,
	)
	if err != nil {
		logger.Error(err, "Error creating SAP share")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error creating SAP share",
			}).
			ErrorLogMessage("Error patching SAP NfsInstance status with error state after failed share creation").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.share = share

	state.ObjAsNfsInstance().SetStateData(StateDataShareId, share.ID)
	state.ObjAsNfsInstance().Status.Id = share.ID

	state.ObjAsNfsInstance().Status.State = "Creating"
	state.ObjAsNfsInstance().Status.CapacityGb = state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb
	if qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", state.ObjAsNfsInstance().Spec.Instance.OpenStack.SizeGb)); err == nil {
		state.ObjAsNfsInstance().Status.Capacity = qty
	}

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error updating SAP NfsInstance state data with created shareId").
		SuccessErrorNil().
		Run(ctx, state)
}
