package cceenfsvolume

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func kcpNfsInstanceCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}
	if state.KcpNfsInstance != nil {
		return nil, ctx
	}

	state.KcpNfsInstance = &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      state.ObjAsCceeNfsVolume().Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      state.ObjAsCceeNfsVolume().Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: state.ObjAsCceeNfsVolume().Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: state.ObjAsCceeNfsVolume().Name,
				Name:      state.ObjAsCceeNfsVolume().Namespace,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.NfsInstanceInfo{
				OpenStack: &cloudcontrolv1beta1.NfsInstanceOpenStack{
					SizeGb: state.ObjAsCceeNfsVolume().Spec.CapacityGb,
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpNfsInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP NfsInstance for CceeNfsVolume", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP NfsInstance for CceeNfsVolume")

	state.ObjAsCceeNfsVolume().Status.State = cloudresourcesv1beta1.StateCreating

	return composed.PatchStatus(state.ObjAsCceeNfsVolume()).
		FailedError(composed.StopWithRequeue).
		ErrorLogMessage("Error patching status for CceeNfsVolume with creating state").
		SuccessErrorNil().
		Run(ctx, state)
}
