package iprange

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preventDeleteOnGcpNfsVolumeUsage(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}
	if state.Provider != nil && *state.Provider != cloudcontrolv1beta1.ProviderGCP {
		// SKR IpRange is NOT GCP, skip check for GcpNfsVolume usage
		logger.WithValues("provider", state.Provider).Info("Skipping preventDeleteOnGcpNfsVolumeUsage.")
		return nil, nil
	}

	gcpNfsVolumesUsingThisIpRange := &cloudresourcesv1beta1.GcpNfsVolumeList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(cloudresourcesv1beta1.IpRangeField, st.Name().String()),
	}
	err := state.Cluster().K8sClient().List(ctx, gcpNfsVolumesUsingThisIpRange, listOps)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing GcpNfsVolumes using IpRange", composed.StopWithRequeue, ctx)
	}

	if len(gcpNfsVolumesUsingThisIpRange.Items) == 0 {
		return nil, nil
	}

	usedByGcpNfsVolumes := fmt.Sprintf("%v", pie.Map(gcpNfsVolumesUsingThisIpRange.Items, func(x cloudresourcesv1beta1.GcpNfsVolume) string {
		return fmt.Sprintf("%s/%s", x.Namespace, x.Name)
	}))

	logger.
		WithValues("usedByGcpNfsVolumes", usedByGcpNfsVolumes).
		Info("IpRange marked for deleting used by GcpNfsVolume")

	state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.StateWarning
	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeWarning,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeDeleteWhileUsed,
			Message: fmt.Sprintf("Can not be deleted while used by: %s", usedByGcpNfsVolumes),
		}).
		ErrorLogMessage("Error updating IpRange status with Warning condition for delete while in use").
		SuccessLogMsg("Forgetting SKR IpRange marked for deleting that is in use").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
