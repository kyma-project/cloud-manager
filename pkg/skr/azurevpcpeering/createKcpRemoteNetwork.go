package azurevpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAzureVpcPeering()

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.RemoteNetwork != nil {
		return nil, nil
	}

	resource, err := util.ParseResourceID(obj.Spec.RemoteVnet)

	if err != nil {
		logger.Error(err, "Error parsing remoteVnet")
		return err, ctx
	}

	remoteNetwork := &cloudcontrolv1beta1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name:      state.ObjAsAzureVpcPeering().Status.Id,
			Namespace: state.KymaRef.Namespace,
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
					Azure: &cloudcontrolv1beta1.AzureNetworkReference{
						SubscriptionId: resource.Subscription,
						ResourceGroup:  resource.ResourceGroup,
						NetworkName:    resource.ResourceName,
					},
				},
			},
			Type: cloudcontrolv1beta1.NetworkTypeExternal,
		},
	}

	err = state.KcpCluster.K8sClient().Create(ctx, remoteNetwork)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP VpcPeering", composed.StopWithRequeue, ctx)
	}

	state.RemoteNetwork = remoteNetwork

	return nil, nil
}
