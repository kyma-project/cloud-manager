package types

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
)

type State interface {
	kcpcommonaction.State

	ObjAsVpcNetwork() *cloudcontrolv1beta1.VpcNetwork
	NormalizedSpecCidrs() []string

	IsKymaTypePredicate(ctx context.Context, st composed.State) bool
	IsGardenerTypePredicate(ctx context.Context, st composed.State) bool
}
