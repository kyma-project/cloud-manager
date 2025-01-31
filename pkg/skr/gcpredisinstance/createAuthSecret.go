package gcpredisinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret != nil {
		return nil, nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   state.Obj().GetNamespace(),
			Name:        getAuthSecretName(state.ObjAsGcpRedisInstance()),
			Labels:      getAuthSecretLabels(state.ObjAsGcpRedisInstance()),
			Annotations: getAuthSecretAnnotations(state.ObjAsGcpRedisInstance()),
			Finalizers: []string{
				api.CommonFinalizerDeletionHook,
			},
		},
		Data: state.GetAuthSecretData(),
	}
	err := state.Cluster().K8sClient().Create(ctx, secret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating secret for GcpRedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("AuthSecret for GcpRedisInstance created")

	return nil, nil
}
