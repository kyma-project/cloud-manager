package scope

import "github.com/kyma-project/cloud-manager/pkg/composed"

func New() composed.Action {
	return composed.ComposeActions(
		"skrCommonScope",
		composed.LoadObj,
		setStatusProcessing,
		loadScope,
		waitScopeReady,
	)
}
