package nfsinstance

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	types2 "github.com/kyma-project/cloud-resources/components/kcp/pkg/nfsinstance/types"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(types2.State)
	logger := composed.LoggerFromCtx(ctx)

	ipRange := &cloudresourcesv1beta1.IpRange{}
	err := state.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      state.ObjAsNfsInstance().Spec.IpRange,
	}, ipRange)

	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading referred IpRange", composed.StopWithRequeue, nil)
	}

	if apierrors.IsNotFound(err) {
		logger.
			WithValues("ipRange", state.ObjAsNfsInstance().Spec.IpRange).
			Error(err, "Referred IpRange does not exist")
		meta.SetStatusCondition(state.ObjAsNfsInstance().Conditions(), metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudresourcesv1beta1.ReasonInvalidIpRangeReference,
			Message: fmt.Sprintf("Referred IpRange %s/%s does not exist", state.Obj().GetNamespace(), state.ObjAsNfsInstance().Spec.IpRange),
		})
		state.ObjAsNfsInstance().Status.State = cloudresourcesv1beta1.ErrorState
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error updating NfsInstance status after referred IpRange not found", composed.StopWithRequeue, nil)
		}

		return composed.StopAndForget, nil
	}

	return nil, nil
}
