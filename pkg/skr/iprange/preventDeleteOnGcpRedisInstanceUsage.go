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

func preventDeleteOnGcpRedisInstanceUsage(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}
	if state.Provider != nil && *state.Provider != cloudcontrolv1beta1.ProviderGCP {
		// SKR IpRange is NOT GCP, skip check for GcpRedisInstance usage
		logger.WithValues("provider", state.Provider).Info("Skipping preventDeleteOnGcpRedisInstanceUsage.")
		return nil, nil
	}
	gcpRedisInstancesUsingThisIpRange := &cloudresourcesv1beta1.GcpRedisInstanceList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudresourcesv1beta1.IpRangeField, st.Name().Name),
	}
	err := state.Cluster().K8sClient().List(ctx, gcpRedisInstancesUsingThisIpRange, listOps)
	if meta.IsNoMatchError(err) {
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing GcpRedisInstances using IpRange", composed.StopWithRequeue, ctx)
	}

	if len(gcpRedisInstancesUsingThisIpRange.Items) == 0 {
		return nil, nil
	}

	usedByGcpRedisInstances := fmt.Sprintf("%v", pie.Map(gcpRedisInstancesUsingThisIpRange.Items, func(x cloudresourcesv1beta1.GcpRedisInstance) string {
		return fmt.Sprintf("%s/%s", x.Namespace, x.Name)
	}))

	logger.
		WithValues("usedByGcpRedisInstances", usedByGcpRedisInstances).
		Info("IpRange marked for deleting used by GcpRedisInstance")

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeWarning,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed,
			Message: fmt.Sprintf("Can not be deleted while used by: %s", usedByGcpRedisInstances),
		}).
		DeriveStateFromConditions(state.MapConditionToState()).
		ErrorLogMessage("Error updating IpRange status with Warning condition for delete while in use").
		SuccessLogMsg("Forgetting SKR IpRange marked for deleting that is in use").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
