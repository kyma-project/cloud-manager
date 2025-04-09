package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
)

func TypeSubnetPredicate(_ context.Context, st composed.State) bool {
	state := st.(iprangetypes.State)

	gcpOptions := state.ObjAsIpRange().Spec.Options.Gcp
	return gcpOptions != nil && gcpOptions.Type == v1beta1.GcpIpRangeTypeSUBNET
}

func TypeGlobalAddressPredicate(_ context.Context, st composed.State) bool {
	state := st.(iprangetypes.State)

	gcpOptions := state.ObjAsIpRange().Spec.Options.Gcp
	return gcpOptions != nil && gcpOptions.Type == v1beta1.GcpIpRangeTypeGLOBAL_ADDRESS
}
