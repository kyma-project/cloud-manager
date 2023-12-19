package iprange

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/abstractions"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/scope"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IPRangeReconciler struct {
	client        client.Client
	eventRecorder record.EventRecorder
	scheme        *runtime.Scheme
	fileReader    abstractions.FileReader
}

func NewIPRangeReconciler(
	client client.Client,
	eventRecorder record.EventRecorder,
	scheme *runtime.Scheme,
	fileReader abstractions.FileReader,
) *IPRangeReconciler {
	return &IPRangeReconciler{
		client:        client,
		eventRecorder: eventRecorder,
		scheme:        scheme,
		fileReader:    fileReader,
	}
}

func (r *IPRangeReconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newState(req.NamespacedName)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *IPRangeReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		scope.New(),
	)
}

func (r *IPRangeReconciler) newState(name types.NamespacedName) *State {
	return newState(
		scope.NewState(
			focal.NewState(
				composed.NewState(
					r.client,
					r.eventRecorder,
					r.scheme,
					name,
					&cloudresourcesv1beta1.IpRange{},
				),
			),
			r.fileReader,
		),
	)
}
