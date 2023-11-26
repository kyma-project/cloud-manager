package genericActions

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func RemoveCondition(conditionType cloudresourcesv1beta1.ConditionType) composed.Action {
	return func(ctx context.Context, state composed.State) error {
		commonStatus := state.Obj().(CommonStatus)
		commonStatus.RemoveCondition(conditionType)
		return nil
	}
}
