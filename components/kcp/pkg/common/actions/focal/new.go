package focal

import "github.com/kyma-project/cloud-resources-manager/components/lib/composed"

func New() composed.Action {
	return composed.ComposeActions(
		"focal",
		loadObj,
		loadScopeFromRef,
		fixInvalidScopeRef,
	)
}
