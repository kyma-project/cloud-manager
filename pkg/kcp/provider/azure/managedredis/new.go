package managedredis

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	managedredistypes "github.com/kyma-project/cloud-manager/pkg/kcp/managedredis/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, err := stateFactory.NewState(ctx, st.(managedredistypes.State), composed.LoggerFromCtx(ctx))
		if err != nil {
			composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap Azure ManagedRedis state")
			obj := st.Obj().(*cloudcontrolv1beta1.AzureManagedRedis)
			obj.Status.State = string(cloudcontrolv1beta1.StateError)
			return composed.UpdateStatus(obj).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
					Message: "Failed to create ManagedRedis state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new Azure ManagedRedis state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"azureManagedRedis",
			actions.AddCommonFinalizer(),
			loadManagedRedis,
			loadDatabase,
			loadPrivateEndPoint,
			loadPrivateDnsZoneGroup,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"azure-managedRedis-create",
					createManagedRedis,
					waitManagedRedisAvailable,
					updateStatusId,
					createDatabase,
					waitDatabaseAvailable,
					createPrivateEndPoint,
					waitPrivateEndPointAvailable,
					createPrivateDnsZoneGroup,
					updateStatus,
				),
				composed.ComposeActions(
					"azure-managedRedis-delete",
					deletePrivateDnsZoneGroup,
					waitPrivateDnsZoneGroupDeleted,
					deletePrivateEndPoint,
					waitPrivateEndPointDeleted,
					deleteDatabase,
					waitDatabaseDeleted,
					deleteManagedRedis,
					waitManagedRedisDeleted,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
