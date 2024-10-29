package iprange

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteOnAzureRedisInstanceUsage(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}
	if state.Provider != nil && *state.Provider != cloudcontrolv1beta1.ProviderAzure {
		return nil, nil
	}
	azureRedisInstancesUsingThisIpRange := &cloudresourcesv1beta1.AzureRedisInstanceList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudresourcesv1beta1.IpRangeField, state.ObjAsIpRange().Name),
	}
	err := state.Cluster().K8sClient().List(ctx, azureRedisInstancesUsingThisIpRange, listOps)
	if meta.IsNoMatchError(err) {
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AzureRedisInstances using IpRange", composed.StopWithRequeue, ctx)
	}

	if len(azureRedisInstancesUsingThisIpRange.Items) == 0 {
		return nil, nil
	}

	usedByAzureRedisInstances := fmt.Sprintf("%v", pie.Map(azureRedisInstancesUsingThisIpRange.Items, func(x cloudresourcesv1beta1.AzureRedisInstance) string {
		return fmt.Sprintf("%s/%s", x.Namespace, x.Name)
	}))

	logger.
		WithValues("usedByAzureRedisInstances", usedByAzureRedisInstances).
		Info("IpRange marked for deleting used by AzureRedisInstance")

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeWarning,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed,
			Message: fmt.Sprintf("Can not be deleted while used by: %s", usedByAzureRedisInstances),
		}).
		DeriveStateFromConditions(state.MapConditionToState()).
		ErrorLogMessage("Error updating IpRange status with Warning condition for delete while in use").
		SuccessLogMsg("Forgetting SKR IpRange marked for deleting that is in use").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
