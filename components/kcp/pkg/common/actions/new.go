package actions

import (
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/scope"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func New() composed.Action {
	return composed.ComposeActions(
		"main",
		focal.New(),
		scope.WhenNoScope(),
	)
}
