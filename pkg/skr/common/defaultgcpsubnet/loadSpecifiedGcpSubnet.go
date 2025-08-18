package defaultgcpsubnet

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadSpecifiedGcpSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	logger := composed.LoggerFromCtx(ctx)

	gcpSubnetRef := state.ObjAsObjWithGcpSubnetRef().GetGcpSubnetRef()
	if len(gcpSubnetRef.Name) == 0 {
		return nil, ctx
	}

	skrGcpSubnet := &cloudresourcesv1beta1.GcpSubnet{}
	err := state.Cluster().K8sClient().Get(ctx, gcpSubnetRef.ObjKey(), skrGcpSubnet)
	if apierrors.IsNotFound(err) {
		logger.Info("Specified SKR GcpSubnet does not exist")
		state.ObjAsObjWithGcpSubnetRef().SetState(cloudresourcesv1beta1.StateError)
		return composed.PatchStatus(state.ObjAsObjWithGcpSubnetRef()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonSubnetNotFound,
				Message: fmt.Sprintf("Specified GcpSubnet %s does not exist", gcpSubnetRef.ObjKey()),
			}).
			SuccessLogMsg(fmt.Sprintf("Forgetting SKR %T after specified GcpSubnet does not exist", state.Obj())).
			ErrorLogMessage(fmt.Sprintf("Error patching SKR %T after specified GcpSubnet does not exist", state.Obj())).
			Run(ctx, state)
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading specified SKR GcpSubnet", composed.StopWithRequeue, ctx)
	}

	state.SetSkrGcpSubnet(skrGcpSubnet)

	return nil, ctx
}
