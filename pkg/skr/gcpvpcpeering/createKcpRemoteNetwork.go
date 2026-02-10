package gcpvpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsGcpVpcPeering()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpRemoteNetwork != nil {
		return nil, nil
	}

	remoteNetwork := &cloudcontrolv1beta1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Status.Id,
			Namespace: state.KymaRef.Namespace,
			Labels: map[string]string{
				common.LabelKymaModule: common.FieldOwner,
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
					Gcp: &cloudcontrolv1beta1.GcpNetworkReference{
						GcpProject:  obj.Spec.RemoteProject,
						NetworkName: obj.Spec.RemoteVpc,
					},
				},
			},
			Type: cloudcontrolv1beta1.NetworkTypeExternal,
		},
	}

	err := state.KcpCluster.K8sClient().Create(ctx, remoteNetwork)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating GCP KCP Remote Network", composed.StopWithRequeue, ctx)
	}

	state.KcpRemoteNetwork = remoteNetwork

	return nil, nil
}
