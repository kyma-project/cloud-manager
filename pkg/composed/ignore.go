package composed

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func ForgetIfIgnored(ctx context.Context, state State) (error, context.Context) {
	_, ok := state.Obj().GetLabels()[cloudcontrolv1beta1.LabelIgnore]
	if ok {
		return StopAndForget, ctx
	}
	return nil, ctx
}
