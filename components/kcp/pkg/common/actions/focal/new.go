package focal

import "github.com/kyma-project/cloud-resources/components/kcp/pkg/common/composed"

func New() composed.Action {
	return composed.ComposeActions(
		"focal",
		loadObj,
		loadScopeFromRef,
		fixInvalidScopeRef,
	)
}
