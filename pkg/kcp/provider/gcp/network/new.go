package network

import (
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(_ StateFactory) composed.Action {
	return composed.ComposeActions(
		"gcpNetwork",
		composed.If(
			composed.MarkedForDeletionPredicate,
			actions.PatchRemoveCommonFinalizer(),
		),
	)
}
