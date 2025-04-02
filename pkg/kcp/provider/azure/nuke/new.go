package nuke

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(nuketypes.State))
		if err != nil {
			err = fmt.Errorf("error creating new Azure Nuke state: %w", err)
			logger.Error(err, "Error")
			obj := st.Obj().(*v1beta1.Nuke)
			return composed.PatchStatus(obj).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonCloudProviderError,
					Message: err.Error(),
				}).
				SuccessError(composed.StopAndForget).
				Run(ctx, st)
		}
		return composed.ComposeActions(
			"azureNuke",
			createAzureClient,
			loadAzureRecoveryVaults,
			loadAzureBackups,
			providerResourceStatusDiscovered,
			deleteAzureBackups,
			deleteAzureVaults,
			providerResourceStatusDeleting,
			providerResourceStatusDeleted,
			checkIfAllProviderResourcesDeleted,
			// continue to parent action
			func(ctx context.Context, state composed.State) (error, context.Context) {
				return nil, ctx
			},
		)(ctx, state)
	}
}
