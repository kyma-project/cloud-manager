package reconciler

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
	ctrl "sigs.k8s.io/controller-runtime"
)

func aggregate(ctx context.Context, state1 composed.LoggableState) (*ctrl.Result, error) {
	state := state1.(*ReconcilingState)
	if state.CloudResources == nil {
		return nil, errors.New("unexpected empty CloudResources")
	}

	sr := state.Obj.(v1beta1.SourceRefAccessor).GetSourceInfo()
	
	state.CloudResources.Spec.Aggregations
}
