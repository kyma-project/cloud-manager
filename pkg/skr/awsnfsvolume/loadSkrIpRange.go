package awsnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadSkrIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	state.SkrIpRange = &cloudresourcesv1beta1.IpRange{}
	err := state.Cluster().K8sClient().Get(ctx, state.ObjAsAwsNfsVolume().Spec.IpRange.ObjKey(), state.SkrIpRange)
	if apierrors.IsNotFound(err) {
		logger.Info("SKR IpRange referred from AwsNfsVolume does not exist")
		return composed.UpdateStatus(state.ObjAsAwsNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonIpRangeNotFound,
				Message: fmt.Sprintf("Specified IpRange %s does not exist", state.ObjAsAwsNfsVolume().Spec.IpRange.ObjKey()),
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady, cloudresourcesv1beta1.ConditionTypeSubmitted).
			Run(ctx, state)
	}
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading SKR IpRange from AwsNfsVolume", composed.StopWithRequeue, ctx)
	}

	logger = logger.WithValues("IpRange", state.ObjAsAwsNfsVolume().Spec.IpRange.ObjKey())
	logger.Info("IpRange loaded")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
