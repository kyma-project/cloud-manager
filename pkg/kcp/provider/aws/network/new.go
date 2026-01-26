package network

import (
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return composed.ComposeActions(
		"awsNetwork",
		composed.If(
			composed.MarkedForDeletionPredicate,
			actions.PatchRemoveCommonFinalizer(),
		),
	)
}
