package managedredis

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func updateDatabase(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedisDatabase == nil {
		return nil, ctx
	}

	// ClientProtocol is hardcoded; this update is a no-op until it becomes user-configurable.
	desiredProtocol := armredisenterprise.ProtocolEncrypted
	currentProtocol := armredisenterprise.ProtocolEncrypted
	if state.managedRedisDatabase.Properties != nil && state.managedRedisDatabase.Properties.ClientProtocol != nil {
		currentProtocol = *state.managedRedisDatabase.Properties.ClientProtocol
	}

	if desiredProtocol == currentProtocol {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Updating Azure Managed Redis database ClientProtocol", "name", obj.Name, "from", currentProtocol, "to", desiredProtocol)

	err := state.client.UpdateDatabase(ctx, state.resourceGroupName, obj.Name, DefaultDatabaseName, armredisenterprise.DatabaseUpdate{
		Properties: &armredisenterprise.DatabaseUpdateProperties{
			ClientProtocol: &desiredProtocol,
		},
	})
	if err != nil {
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to update Azure Managed Redis database: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
