package runtime

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/rate"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func runtimeAddHandledAnnotation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if _, ok := state.ObjAsRuntime().Labels[cloudcontrolv1beta1.AnnotationRuntimeHandled]; ok {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Annotating Runtime as handled")

	_, err := composed.PatchObjMergeAnnotation(ctx, cloudcontrolv1beta1.AnnotationRuntimeHandled, "true", state.ObjAsRuntime(), state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed to add handled annotation to runtime", composed.StopWithRequeueDelay(rate.Quick.When(state.ObjAsRuntime())), ctx)
	}

	return nil, ctx
}
