package alicloudnfsvolume

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR AlicloudNfsVolume is marked for deletion, do not create mirror in KCP
		return nil, nil
	}

	if state.KcpNfsInstance != nil {
		// mirror NfsInstance in KCP is already created
		return nil, nil
	}

	state.KcpNfsInstance = &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      state.ObjAsAlicloudNfsVolume().Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      state.ObjAsAlicloudNfsVolume().Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: state.ObjAsAlicloudNfsVolume().Namespace,
				common.LabelKymaModule:                   common.FieldOwner,
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: state.ObjAsAlicloudNfsVolume().Namespace,
				Name:      state.ObjAsAlicloudNfsVolume().Name,
			},
			IpRange: cloudcontrolv1beta1.IpRangeRef{
				Name: state.SkrIpRange.Status.Id,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Instance: cloudcontrolv1beta1.NfsInstanceInfo{
				Alicloud: &cloudcontrolv1beta1.NfsInstanceAlicloud{
					// Capacity is not propagated to KCP - AliCloud NAS is elastic.
					StorageType:  cloudcontrolv1beta1.AlicloudStorageType(state.ObjAsAlicloudNfsVolume().Spec.StorageType),
					ProtocolType: cloudcontrolv1beta1.AlicloudProtocolTypeNFS,
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpNfsInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP NfsInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP NfsInstance")

	//Set the state to creating
	state.ObjAsAlicloudNfsVolume().Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(state.ObjAsAlicloudNfsVolume()).
		ErrorLogMessage("Error setting Creating state on AlicloudNfsVolume").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
