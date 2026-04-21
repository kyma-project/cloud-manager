package runtime

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type State struct {
	composed.State

	Subscription *cloudcontrolv1beta1.Subscription
}
