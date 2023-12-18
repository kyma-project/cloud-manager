package scope

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/util"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func loadKyma(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(State)

	kymaUnstructured := util.NewKymaUnstructured()
	err := state.Client().Get(ctx, types.NamespacedName{
		Name:      state.CommonObj().KymaName(),
		Namespace: state.Obj().GetNamespace(),
	}, kymaUnstructured)
	if apierrors.IsNotFound(err) {
		meta.SetStatusCondition(state.CommonObj().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ReasonInvalidKymaName,
			Message: fmt.Sprintf("The Kyma CR '%s' does not exit", state.CommonObj().KymaName()),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			err = fmt.Errorf("error updating status reason invalid kyma: %w", err)
			logger.Error(err, "Error loading Kyma")
			return composed.StopWithRequeue, nil // requeue so object status is updated to invalid kyma
		}

		return composed.StopAndForget, nil // status is set, no requeue since it refers to missing kyma cr
	}

	if err != nil {
		err = fmt.Errorf("error loading Kyma CR: %w", err)
		logger.Error(err, "Error")
		return composed.StopWithRequeue, nil // requeue, try again
	}

	// Kyma CR is loaded, read the shootName now

	state.SetShootName(kymaUnstructured.GetLabels()["kyma-project.io/shoot-name"])

	logger = logger.WithValues("shootName", state.ShootName)
	logger.Info("Shoot name found")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
