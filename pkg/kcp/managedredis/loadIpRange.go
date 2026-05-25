package managedredis

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	managedredistypes "github.com/kyma-project/cloud-manager/pkg/kcp/managedredis/types"
)

// loadIpRange validates that the referenced IpRange exists and is accessible.
// The IpRange field is optional; if unset the action is a no-op.
// Note: the Azure provider derives subnet IDs directly from the Scope VPC network
// and does not consume IpRange.Spec.Network — this action exists solely for
// existence/accessibility validation so users get a clear error on misconfiguration.
func loadIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(managedredistypes.State)
	logger := composed.LoggerFromCtx(ctx)

	obj := state.ObjAsAzureManagedRedis()
	ipRangeName := obj.Spec.IpRange.Name

	if ipRangeName == "" {
		return nil, ctx
	}

	ipRange := &cloudcontrolv1beta1.IpRange{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      ipRangeName,
	}, ipRange)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading referred IpRange", composed.StopWithRequeue, ctx)
	}

	if apierrors.IsNotFound(err) {
		logger.
			WithValues("ipRange", ipRangeName).
			Error(err, "Referred IpRange does not exist")
		meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonInvalidIpRangeReference,
			Message: fmt.Sprintf("Referred IpRange %s/%s does not exist", state.Obj().GetNamespace(), ipRangeName),
		})
		obj.SetState(string(cloudcontrolv1beta1.StateError))
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating AzureManagedRedis status after referred IpRange not found", composed.StopWithRequeue, ctx)
		}

		return composed.StopAndForget, nil
	}

	return nil, ctx
}
