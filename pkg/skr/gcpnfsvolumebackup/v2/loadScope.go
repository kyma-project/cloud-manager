package v2

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
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	backup := state.ObjAsGcpNfsVolumeBackup()

	logger = logger.WithValues(
		"scope", state.KymaRef.Name,
		"scopeNamespace", state.KymaRef.Namespace,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	scope := &cloudcontrolv1beta1.Scope{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Name:      state.KymaRef.Name,
		Namespace: state.KymaRef.Namespace,
	}, scope)

	if apierrors.IsNotFound(err) {
		logger.Info("Scope not found")

		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonScopeNotFound,
				Message: fmt.Sprintf("Scope %s does not exist", state.KymaRef.Name),
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Scope", composed.StopWithRequeue, ctx)
	}

	logger = logger.WithValues(
		"provider", scope.Spec.Provider,
		"region", scope.Spec.Region,
		"shootName", scope.Spec.ShootName,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.Scope = scope

	return nil, ctx
}
