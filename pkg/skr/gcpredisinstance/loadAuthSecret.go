package gcpredisinstance

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func loadAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	gcpRedisInstance := state.ObjAsGcpRedisInstance()

	secret := &corev1.Secret{}
	authSecretName := getAuthSecretName(state.ObjAsGcpRedisInstance())
	err := state.Cluster().K8sClient().Get(ctx, types.NamespacedName{
		Namespace: state.Obj().GetNamespace(),
		Name:      authSecretName,
	}, secret)

	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, nil
		}
		return composed.LogErrorAndReturn(err, "Error getting Secret by getAuthSecretName()", composed.StopWithRequeue, ctx)
	}

	if secret.Labels[cloudresourcesv1beta1.LabelRedisInstanceId] != gcpRedisInstance.Status.Id {
		gcpRedisInstance.Status.State = cloudresourcesv1beta1.StateError
		errMsg := fmt.Sprintf("Auth secret %s belongs to another resource", authSecretName)
		return composed.UpdateStatus(gcpRedisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage(errMsg).
			SuccessLogMsg("Updated and forgot SKR GcpRedisInstance status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	state.AuthSecret = secret

	return nil, nil
}
