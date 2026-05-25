package managedredis

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedis == nil || state.managedRedis.Properties == nil {
		return nil, ctx
	}

	hostname := ""
	if state.managedRedis.Properties.HostName != nil {
		hostname = *state.managedRedis.Properties.HostName
	}

	if obj.Status.State == string(cloudcontrolv1beta1.StateReady) &&
		obj.Status.ObservedGeneration == obj.Generation &&
		obj.Status.PrimaryEndpoint == hostname &&
		obj.Status.Port == RedisPort {
		return nil, ctx
	}

	keys, err := state.client.ListKeys(ctx, state.resourceGroupName, obj.Name, DefaultDatabaseName)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error retrieving Azure Managed Redis access keys", composed.StopWithRequeue, ctx)
	}
	authString := ""
	if keys != nil && keys.PrimaryKey != nil {
		authString = *keys.PrimaryKey
	}

	obj.Status.State = string(cloudcontrolv1beta1.StateReady)
	obj.Status.ObservedGeneration = obj.Generation
	obj.Status.PrimaryEndpoint = hostname
	obj.Status.Port = RedisPort
	obj.Status.AuthString = authString
	return composed.UpdateStatus(obj).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Azure Managed Redis is ready",
		}).
		ErrorLogMessage("Error updating KCP AzureManagedRedis status after setting Ready condition").
		SuccessLogMsg("KCP AzureManagedRedis is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, st)
}
