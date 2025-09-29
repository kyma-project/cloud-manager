package sapnfsvolume

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
			Name:      state.ObjAsSapNfsVolume().Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      state.ObjAsSapNfsVolume().Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: state.ObjAsSapNfsVolume().Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Name,
			},
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: state.ObjAsSapNfsVolume().Name,
				Name:      state.ObjAsSapNfsVolume().Namespace,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.NfsInstanceInfo{
				OpenStack: &cloudcontrolv1beta1.NfsInstanceOpenStack{
					SizeGb: state.ObjAsSapNfsVolume().Spec.CapacityGb,
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpNfsInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP NfsInstance for SapNfsVolume", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP NfsInstance for SapNfsVolume")

	state.ObjAsSapNfsVolume().Status.State = cloudresourcesv1beta1.StateCreating

	return composed.PatchStatus(state.ObjAsSapNfsVolume()).
		FailedError(composed.StopWithRequeue).
		ErrorLogMessage("Error patching status for SapNfsVolume with creating state").
		SuccessErrorNil().
		Run(ctx, state)
}
