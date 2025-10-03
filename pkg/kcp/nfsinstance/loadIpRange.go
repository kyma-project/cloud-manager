package nfsinstance

import (
	"context"
	"fmt"

	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(nfsinstancetypes.State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := &cloudcontrolv1beta1.IpRange{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      state.ObjAsNfsInstance().Spec.IpRange.Name,
	}, ipRange)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading referred IpRange", composed.StopWithRequeue, ctx)
	}

	if apierrors.IsNotFound(err) {
		logger.
			WithValues("ipRange", state.ObjAsNfsInstance().Spec.IpRange.Name).
			Error(err, "Referred IpRange does not exist")
		meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonInvalidIpRangeReference,
			Message: fmt.Sprintf("Referred IpRange %s/%s does not exist", state.Obj().GetNamespace(), state.ObjAsNfsInstance().Spec.IpRange.Name),
		})
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.StateError
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating NfsInstance status after referred IpRange not found", composed.StopWithRequeue, ctx)
		}

		return composed.StopAndForget, nil
	}

	//Store IPRange into the state object.
	state.SetIpRange(ipRange)

	return nil, nil
}
