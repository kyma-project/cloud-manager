package v2

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// New creates the main reconciliation action for GCP NfsInstance v2.
// It composes all necessary actions in the correct sequence.
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		logger.Info("Creating GCP NfsInstance v2 state")

		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			logger.Error(err, "Error creating state")
			nfsInstance := st.Obj().(*v1beta1.NfsInstance)
			return composed.UpdateStatus(nfsInstance).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: fmt.Sprintf("Failed to initialize GCP client: %s", err.Error()),
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new GCP NfsInstance v2 state: %s", err)).
				Run(ctx, st)
		}

		return composeActions()(ctx, state)
	}
}
