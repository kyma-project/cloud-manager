package awsvpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func createKcpVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpVpcPeering != nil {
		return nil, nil
	}

	obj := state.ObjAsAwsVpcPeering()

	state.KcpVpcPeering = &cloudcontrolv1beta1.VpcPeering{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      obj.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: obj.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.VpcPeeringSpec{
			RemoteRef: cloudcontrolv1beta1.RemoteRef{
				Namespace: obj.Namespace,
				Name:      obj.Name,
			},
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},

			Details: &cloudcontrolv1beta1.VpcPeeringDetails{
				DeleteRemotePeering: obj.Spec.DeleteRemotePeering,
				RemoteNetwork: klog.ObjectRef{
					Name:      state.RemoteNetwork.Name,
					Namespace: state.RemoteNetwork.Namespace,
				},
				LocalNetwork: klog.ObjectRef{
					Name:      common.KcpNetworkKymaCommonName(state.KymaRef.Name),
					Namespace: state.KymaRef.Namespace,
				},
			},
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, state.KcpVpcPeering)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP VpcPeering", composed.StopWithRequeue, ctx)
	}

	logger.Info("Created KCP VpcPeering", "id", obj.Status.Id)

	return nil, nil
}
