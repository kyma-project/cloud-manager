package awsredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
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
			Name:        getAuthSecretName(state.ObjAsAwsRedisInstance()),
			Labels:      getAuthSecretLabels(state.ObjAsAwsRedisInstance()),
			Annotations: getAuthSecretAnnotations(state.ObjAsAwsRedisInstance()),
			Finalizers: []string{
				v1beta1.Finalizer,
			},
		},
		Data: state.GetAuthSecretData(),
	}
	err := state.Cluster().K8sClient().Create(ctx, secret)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating secret for AwsRedisInstance", composed.StopWithRequeue, ctx)
	}

	logger.Info("AuthSecret for AwsRedisInstance created")

	return nil, nil
}
