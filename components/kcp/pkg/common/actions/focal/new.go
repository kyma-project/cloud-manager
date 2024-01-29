package focal

import "github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"

func New() composed.Action {
	return composed.ComposeActions(
		"focal",
		composed.LoadObj,
		loadScopeFromRef,
	)
}
