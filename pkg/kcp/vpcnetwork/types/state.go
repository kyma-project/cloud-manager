package types

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
)

type State interface {
	kcpcommonaction.State

	ObjAsVpcNetwork() *cloudcontrolv1beta1.VpcNetwork
	NormalizedSpecCidrs() []string
}
