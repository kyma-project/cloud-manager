package awsvpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsAwsVpcPeering()

	remoteNetwork := &cloudcontrolv1beta1.Network{}

	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      obj.Status.Id,
	}, remoteNetwork)

	if apierrors.IsNotFound(err) {
		state.RemoteNetwork = nil
		logger.Info("KCP Network does not exist")
		return nil, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP RemoteNetwork", composed.StopWithRequeue, ctx)
	}

	state.RemoteNetwork = remoteNetwork

	return nil, nil
}
