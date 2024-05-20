package gcpnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func validatePV(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	nfsVolume := state.ObjAsGcpNfsVolume()

	//If marked for deletion, continue without validation.
	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//Get the PV with the desired name
	pvName := getVolumeName(nfsVolume)
	pv := &corev1.PersistentVolume{}
	err := state.SkrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: pvName}, pv)

	//If PV does not exist, continue.
	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	//If some other error, stop and requeue
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting PersistentVolume by name", composed.StopWithRequeue, ctx)
	}

	//If PV belongs to this NfsVolume, continue
	name, nmExists := pv.Labels[cloudresourcesv1beta1.LabelNfsVolName]
	ns, nsExists := pv.Labels[cloudresourcesv1beta1.LabelNfsVolNS]
	if nmExists && nsExists && name == nfsVolume.Name && ns == nfsVolume.Namespace {
		return nil, nil
	}

	//If PV exists with the same name, but does not belong to this NfsVolume, then mark it as an error
	state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
	return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonPVNameInvalid,
			Message: fmt.Sprintf("PV with the name %s already exists.", pvName),
		}).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
		ErrorLogMessage("Error updating GcpNfsVolume status for invalid PV name").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
