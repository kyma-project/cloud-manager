package scope

import "github.com/kyma-project/cloud-manager/pkg/composed"

// New creates common scope flow action, as the package name says. Do not
// crete alternate flows, since then it's not common, it's specific to what
// you're doing. If it doesn't fit, then probably you're doing something
// wrong.
func New() composed.Action {
	return composed.ComposeActionsNoName(
		composed.LoadObj,
		statusUpdateProcessingLegacy,
		statusProcessing,
		loadScope,
		waitScopeReadyLegacy,
		waitScopeReady,
	)
}
