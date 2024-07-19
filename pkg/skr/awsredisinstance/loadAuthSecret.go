package awsredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	secret := &corev1.Secret{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      getAuthSecretName(state.ObjAsAwsRedisInstance()),
	}, secret)
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error getting Secret by getAuthSecretName()", composed.StopWithRequeue, ctx)
	}

	if err == nil {
		state.AuthSecret = secret
	}

	return nil, nil
}
