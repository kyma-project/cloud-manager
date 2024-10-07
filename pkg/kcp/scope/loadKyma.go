package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadKyma(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	kymaUnstructured := util.NewKymaUnstructured()
	err := state.Cluster().K8sClient().Get(ctx, state.Name(), kymaUnstructured)

	if apierrors.IsNotFound(err) {
		logger.Info("Kyma CR does not exist")

		if !NukeScopesWithoutKyma {
			return composed.StopAndForget, nil
		}

		nuke := &cloudcontrolv1beta1.Nuke{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: state.Obj().GetNamespace(),
				Name:      state.Obj().GetName(),
			},
			Spec: cloudcontrolv1beta1.NukeSpec{
				Scope: cloudcontrolv1beta1.ScopeRef{
					Name: state.Obj().GetName(),
				},
			},
		}

		err = state.Cluster().K8sClient().Create(ctx, nuke)
		if apierrors.IsAlreadyExists(err) {
			logger.Info("Nuke for this Scope already exists")
			return composed.StopAndForget, nil
		}
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error creating Nuke", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}

		logger.Info("Nuke created")

		return composed.StopAndForget, nil
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error loading Kyma CR", composed.StopWithRequeue, ctx)
	}

	state.kyma = kymaUnstructured

	return nil, nil
}
