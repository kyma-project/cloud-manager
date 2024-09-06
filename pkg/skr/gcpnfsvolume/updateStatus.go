package gcpnfsvolume

import (
	"context"
	"fmt"
	"strconv"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	logger := composed.LoggerFromCtx(ctx)
	logger.
		WithValues("kcpNfsInstanceConditions", pie.Map(state.KcpNfsInstance.Status.Conditions, func(c metav1.Condition) string {
			return fmt.Sprintf("%s:%s", c.Type, c.Status)
		})).
		Info("Updating SKR GcpNfsVolume status from KCP NfsInstance conditions")

	kcpCondErr := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	kcpCondReady := meta.FindStatusCondition(state.KcpNfsInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

	skrCondErr := meta.FindStatusCondition(state.ObjAsGcpNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
	skrCondReady := meta.FindStatusCondition(state.ObjAsGcpNfsVolume().Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	if kcpCondErr != nil && skrCondErr == nil {
		logger.Info("Updating SKR GcpNfsVolume status with Error condition")
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.PatchStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: kcpCondErr.Message,
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with Error condition due to KCP error").
			SuccessError(composed.StopAndForget). // do not continue further with the flow
			Run(ctx, state)
	}
	if kcpCondErr != nil && skrCondErr != nil {
		// already with Error condition
		return composed.StopAndForget, nil
	}

	if kcpCondReady != nil && skrCondReady == nil {
		logger.Info("Updating SKR GcpNfsVolume status with Ready condition")
		state.ObjAsGcpNfsVolume().Status.CapacityGb = state.KcpNfsInstance.Status.CapacityGb
		state.ObjAsGcpNfsVolume().Status.Hosts = state.KcpNfsInstance.Status.Hosts
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeReady
		return composed.PatchStatus(state.ObjAsGcpNfsVolume()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: kcpCondReady.Message,
			}).
			ErrorLogMessage("Error updating GcpNfsVolume status with Ready condition").
			SuccessLogMsg("GcpNfsVolume status got updated with Ready condition "+strconv.Itoa(state.KcpNfsInstance.Spec.Instance.Gcp.CapacityGb)).
			SuccessError(composed.StopWithRequeue). // have to continue and requeue to create PV
			Run(ctx, state)
	}
	if kcpCondReady != nil && skrCondReady != nil {
		// already with Ready condition
		// continue with next actions to create PV
		return nil, nil
	}

	// no conditions on KCP NfsInstance
	// keep looping until KCP NfsInstance gets some condition
	switch state.KcpNfsInstance.Status.State {
	case client.Creating:
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeCreating
	case client.Updating:
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeUpdating
	case client.Deleting:
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeDeleting
	default:
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeProcessing

	}
	return composed.PatchStatus(state.ObjAsGcpNfsVolume()).
		SuccessError(composed.StopWithRequeueDelay(2*util.Timing.T100ms())).
		Run(ctx, state)
}
