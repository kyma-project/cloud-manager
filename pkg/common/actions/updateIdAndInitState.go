package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func UpdateIdAndInitState(statusState string) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		if composed.MarkedForDeletionPredicate(ctx, state) {
			return nil, nil
		}

		name := state.Obj().GetObjectKind().GroupVersionKind().GroupKind().String()

		obj, ok := state.Obj().(composed.ObjWithStatusId)

		if !ok {
			return composed.LogErrorAndReturn(common.ErrLogical,
				fmt.Sprintf("%T is not of type composed.ObjWithStatusId", state.Obj()),
				composed.StopAndForget,
				ctx)
		}

		if obj.Id() != "" {
			return nil, ctx
		}

		id := uuid.NewString()

		if state.Obj().GetLabels() == nil {
			state.Obj().SetLabels(make(map[string]string))
		}

		state.Obj().GetLabels()[cloudresourcesv1beta1.LabelId] = id
		state.Obj().GetLabels()[common.LabelKymaModule] = "cloud-manager"

		err := state.UpdateObj(ctx)

		if err != nil {
			return composed.LogErrorAndReturn(err, fmt.Sprintf("Error updating %s with ID label", name), composed.StopWithRequeue, ctx)
		}

		logger.Info(fmt.Sprintf("%s updated with ID label", name))

		obj.SetId(id)

		obj.SetState(statusState)

		err = state.UpdateObjStatus(ctx)

		if err != nil {
			return composed.LogErrorAndReturn(err, fmt.Sprintf("Error updating %s status with ID", name), composed.StopWithRequeue, ctx)
		}

		logger.Info(fmt.Sprintf("%s updated with ID status", name))

		return composed.StopWithRequeueDelay(util.Timing.T100ms()), nil
	}
}
