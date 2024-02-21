package awsnfsvolume

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func loadKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsAwsNfsVolume().Status.Id == "" {
		return composed.LogErrorAndReturn(
			errors.New("missing SKR AwsNfsVolume state.id"),
			"Logical error in loadKcpNfsInstance",
			composed.StopAndForget,
			ctx,
		)
	}

	kcpNfsInstnace := &cloudcontrolv1beta1.NfsInstance{}
	err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.KymaRef.Namespace,
		Name:      state.ObjAsAwsNfsVolume().Status.Id,
	}, kcpNfsInstnace)
	if apierrors.IsNotFound(err) {
		logger.Info("KCP NfsInstance does not exist")
		return nil, nil
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading KCP NfsInstance", composed.StopWithRequeue, ctx)
	}

	state.KcpNfsInstance = kcpNfsInstnace

	return nil, nil
}
