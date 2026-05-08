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

func createDatabase(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedisDatabase != nil {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Creating Azure Managed Redis database", "name", obj.Name)

	clientProtocol := armredisenterprise.ProtocolEncrypted
	accessKeysAuth := armredisenterprise.AccessKeysAuthenticationEnabled
	clusteringPolicy := armredisenterprise.ClusteringPolicy(obj.Spec.ClusteringPolicy)
	db := armredisenterprise.Database{
		Properties: &armredisenterprise.DatabaseCreateProperties{
			ClusteringPolicy:         &clusteringPolicy,
			ClientProtocol:           &clientProtocol,
			AccessKeysAuthentication: &accessKeysAuth,
		},
	}

	err := state.client.CreateOrUpdateDatabase(ctx, state.resourceGroupName, obj.Name, DefaultDatabaseName, db)
	if err != nil {
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.UpdateStatus(obj).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: fmt.Sprintf("Failed to create Azure Managed Redis database: %s", err),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, st)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
