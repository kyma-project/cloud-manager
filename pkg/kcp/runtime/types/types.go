package types

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
)

type State interface {
	composed.State
	ObjAsRuntime() *infrastructuremanagerv1.Runtime
}
