package scope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func nukeCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !cloudcontrolv1beta1.AutomaticNuke {
		return nil, ctx
	}

	if state.nuke != nil {
		return nil, ctx
	}

	logger := ctrl.LoggerFrom(ctx)
	logger.Info("Creating Nuke")

	// Take care! Scope might not exist, thus it's easier just to use the reconciliation request

	nuke := &cloudcontrolv1beta1.Nuke{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: state.Name().Namespace,
			Name:      state.Name().Name,
		},
		Spec: cloudcontrolv1beta1.NukeSpec{
			Scope: cloudcontrolv1beta1.ScopeRef{
				Name: state.Name().Name,
			},
		},
	}

	err := state.Cluster().K8sClient().Create(ctx, nuke)
	if apierrors.IsAlreadyExists(err) {
		logger.Info("Nuke for this Scope already exists")
		return composed.StopWithRequeue, ctx
	}
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating Nuke", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	logger.Info("Nuke created")

	return composed.StopWithRequeue, ctx
}
