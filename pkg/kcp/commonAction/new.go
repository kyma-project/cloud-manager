package commonAction

import "github.com/kyma-project/cloud-manager/pkg/composed"

func New() composed.Action {
	return composed.ComposeActionsNoName(
		composed.LoadObj,
		subscriptionLoad,
		// TODO: setup feature context
	)
}
