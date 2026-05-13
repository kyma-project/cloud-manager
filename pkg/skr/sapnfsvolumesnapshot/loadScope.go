package sapnfsvolumesnapshot

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func loadScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      state.KymaRef.Name,
		Namespace: state.KymaRef.Namespace,
	}, scope)

	if apierrors.IsNotFound(err) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(err, "Scope not found", "scope", state.KymaRef.Name)
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonMissingScope,
				Message: fmt.Sprintf("Scope %s does not exist", state.KymaRef.Name),
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope", composed.StopWithRequeue, ctx)
	}

	state.Scope = scope

	return nil, ctx
}
