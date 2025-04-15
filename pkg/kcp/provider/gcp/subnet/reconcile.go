package subnet

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	ctrl "sigs.k8s.io/controller-runtime"
)

type GcpSubnetReconciler interface {
	reconcile.Reconciler
}

type gcpSubnetReconciler struct {
	composedStateFactory composed.StateFactory
	focalStateFactory    focal.StateFactory

	stateFactory StateFactory
}

func NewGcpSubnetReconciler(
	composedStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	stateFactory StateFactory,
) GcpSubnetReconciler {
	return &gcpSubnetReconciler{
		composedStateFactory: composedStateFactory,
		focalStateFactory:    focalStateFactory,
		stateFactory:         stateFactory,
	}
}

func (r *gcpSubnetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.newFocalState(req.NamespacedName)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("kcpgcpsubnet", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *gcpSubnetReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpSubnet{}),
		focal.New(),
		r.newFlow(),
	)
}

func (r *gcpSubnetReconciler) newFlow() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, err := r.stateFactory.NewState(ctx, st.(focal.State))
		if err != nil {
			composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap GCP Subnet state")
			subnet := st.Obj().(*cloudcontrolv1beta1.GcpSubnet)
			subnet.Status.State = cloudcontrolv1beta1.StateError
			return composed.UpdateStatus(subnet).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
					Message: "Failed to create Subnet state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new GCP Subnet state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"privateSubnet",
			actions.AddCommonFinalizer(),
			loadSubnet,
			loadConnectionPolicy,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"privateSubnet-create",
					createSubnet,
					copyCidrToStatus,
					createConnectionPolicy,
					updateStatusId,
					updateStatus,
				),
				composed.ComposeActions(
					"privateSubnet-delete",
					removeReadyCondition,
					preventDeleteOnGcpRedisClusterUsage,
					deleteConnectionPolicy,
					deleteSubnet,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}

func (r *gcpSubnetReconciler) newFocalState(name types.NamespacedName) focal.State {
	return r.focalStateFactory.NewState(
		r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.GcpSubnet{}),
	)
}
