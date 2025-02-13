package cceenfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func pvcValidate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	desiredPvcName := state.ObjAsCceeNfsVolume().GetPVCName()
	pvc := &corev1.PersistentVolumeClaim{}
	err := state.Cluster().K8sClient().Get(ctx, client.ObjectKey{Namespace: state.ObjAsCceeNfsVolume().Namespace, Name: desiredPvcName}, pvc)

	if apierrors.IsNotFound(err) {
		return nil, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting PVC to validated CceeNfsVolume PVC", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	parentName, nameLabelExists := pvc.Labels[cloudresourcesv1beta1.LabelNfsVolName]
	parentNamespace, namespaceLabelExists := pvc.Labels[cloudresourcesv1beta1.LabelNfsVolNS]
	if nameLabelExists &&
		namespaceLabelExists &&
		parentName == state.ObjAsCceeNfsVolume().Name &&
		parentNamespace == state.ObjAsCceeNfsVolume().Namespace {
		return nil, ctx
	}

	state.ObjAsCceeNfsVolume().Status.State = cloudresourcesv1beta1.StateError
	return composed.PatchStatus(state.ObjAsCceeNfsVolume()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonPVNameInvalid,
			Message: fmt.Sprintf("Desired PVC name %s already exists with different owner", desiredPvcName),
		}).
		FailedError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		ErrorLogMessage("Error patching CceeNfsVolume status with error condition when PVC already exists with different owner").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
