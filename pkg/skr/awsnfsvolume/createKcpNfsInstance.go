package awsnfsvolume

import (
	"context"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is marked for deletion, do not create mirror in KCP
		return nil, nil
	}

	if state.KcpNfsInstance != nil {
		// mirror IpRange in KCP is already created
		return nil, nil
	}

	state.KcpNfsInstance = &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.NewString(),
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      state.Name().Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: state.Name().Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.NfsInstanceSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{},
			IpRange:   cloudcontrolv1beta1.IpRangeRef{},
			Scope:     cloudcontrolv1beta1.ScopeRef{},
			Instance:  cloudcontrolv1beta1.NfsInstanceInfo{},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpNfsInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP NfsInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP NfsInstance")

	return nil, nil
}
