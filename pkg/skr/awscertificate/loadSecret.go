package awscertificate

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsAwsCertificate()

	secret := &corev1.Secret{}
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: cert.Spec.SecretRef.Namespace,
		Name:      cert.Spec.SecretRef.Name,
	}, secret)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Secret not found")
			return composed.NewStatusPatcherComposed(cert).
				MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
					c.SetStatusProviderError("Secret not found")
				}).
				OnSuccess(composed.Requeue).
				Run(ctx, state.Cluster().K8sClient())
		}
		return composed.LogErrorAndReturn(err, "Error loading Secret", composed.StopWithRequeue, ctx)
	}

	state.secret = secret
	logger.Info("Secret loaded successfully")

	return nil, ctx
}
