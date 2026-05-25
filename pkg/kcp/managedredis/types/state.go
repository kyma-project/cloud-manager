package types

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
)

type State interface {
	focal.State
	ObjAsAzureManagedRedis() *cloudcontrolv1beta1.AzureManagedRedis
}
