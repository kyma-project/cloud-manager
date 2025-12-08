package awsvpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAwsVpcPeering()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.RemoteNetwork != nil {
		return nil, nil
	}

	remoteNetwork := &cloudcontrolv1beta1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: "cloud-manager",
			},
			Annotations: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
				cloudcontrolv1beta1.LabelRemoteName:      obj.Name,
				cloudcontrolv1beta1.LabelRemoteNamespace: obj.Namespace,
			},
		},
		Spec: cloudcontrolv1beta1.NetworkSpec{
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.KymaRef.Name,
			},
			Network: cloudcontrolv1beta1.NetworkInfo{
				Reference: &cloudcontrolv1beta1.NetworkReference{
					Aws: &cloudcontrolv1beta1.AwsNetworkReference{
						Region:       obj.Spec.RemoteRegion,
						AwsAccountId: obj.Spec.RemoteAccountId,
						VpcId:        obj.Spec.RemoteVpcId,
					},
				},
			},
			Type: cloudcontrolv1beta1.NetworkTypeExternal,
		},
	}

	logger.Info("Remote network reference created")

	err := state.KcpCluster.K8sClient().Create(ctx, remoteNetwork)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP VpcPeering", composed.StopWithRequeue, ctx)
	}

	state.RemoteNetwork = remoteNetwork

	return nil, nil
}
