package defaultiprange

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadSpecifiedIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	logger := composed.LoggerFromCtx(ctx)

	ipRangeRef := state.ObjAsObjWithIpRangeRef().GetIpRangeRef()
	if len(ipRangeRef.Name) == 0 {
		return nil, ctx
	}

	skrIpRange := &cloudresourcesv1beta1.IpRange{}
	err := state.Cluster().K8sClient().Get(ctx, ipRangeRef.ObjKey(), skrIpRange)
	if apierrors.IsNotFound(err) {
		logger.Info("Specified SKR IpRange does not exist")
		state.ObjAsObjWithIpRangeRef().SetState(cloudresourcesv1beta1.StateError)
		return composed.PatchStatus(state.ObjAsObjWithIpRangeRef()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeNotFound,
				Message: fmt.Sprintf("Specified IpRange %s does not exist", ipRangeRef.ObjKey()),
			}).
			SuccessLogMsg(fmt.Sprintf("Forgetting SKR %T after specified IpRange does not exist", state.Obj())).
			ErrorLogMessage(fmt.Sprintf("Error patching SKR %T after specified IpRange does not exist", state.Obj())).
			Run(ctx, state)
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading specified SKR IpRange", composed.StopWithRequeue, ctx)
	}

	state.SetSkrIpRange(skrIpRange)

	return nil, ctx
}
