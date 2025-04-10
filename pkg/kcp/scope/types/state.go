package types

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type State interface {
	composed.State
	ObjAsScope() *cloudcontrolv1beta1.Scope
	ExposedData() *ExposedData
}

type ExposedData struct {
	NatGatewayIps []string
}
