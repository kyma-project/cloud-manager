package crgcpnfsvolume

import (
	"context"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is marked for deletion, do not create mirror in KCP
		return nil, nil
	}

	if state.KcpNfsInstance != nil {
		// mirror NfsInstance in KCP is already created
		return nil, nil
	}

	state.KcpNfsInstance = &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewString(),
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				labelKymaName:        state.KymaRef.Name,
				labelRemoteName:      state.Name().Name,
				labelRemoteNamespace: state.Name().Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: state.ObjAsGcpNfsVolume().Namespace,
				Name:      state.ObjAsGcpNfsVolume().Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.ObjAsGcpNfsVolume().Spec.IpRange.Name,
			},
			Instance: cloudcontrolv1beta1.NfsInstanceInfo{
				Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
					Location:      state.ObjAsGcpNfsVolume().Spec.Location,
					Tier:          cloudcontrolv1beta1.GcpFileTier(state.ObjAsGcpNfsVolume().Spec.Tier),
					FileShareName: state.ObjAsGcpNfsVolume().Spec.FileShareName,
					CapacityGb:    state.ObjAsGcpNfsVolume().Spec.CapacityGb,
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpNfsInstance)
	if err != nil {
		logger.Error(err, "Error creating KCP NfsInstance")
		return composed.StopWithRequeue, nil
	}
	logger.
		WithValues("kcpNfsInstanceName", state.KcpNfsInstance.Name).
		Info("KCP NFS instance created")

	return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeSubmitted,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonSubmissionSucceeded,
			Message: "Resource is submitted for provisioning",
		}).
		ErrorLogMessage("Error updating GcpNfsVolume status with submitted condition").
		Run(ctx, state)
}
