package v2

import (
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// composeActions creates the main action pipeline for GCP NfsInstance reconciliation.
func composeActions() composed.Action {
	return composed.ComposeActions(
		"gcpNfsInstanceV2",
		actions.AddCommonFinalizer(),

		pollOperation,

		loadInstance,

		composed.IfElse(
			composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"create-update",
				createInstance,
				waitInstanceReady,
				modifyCapacityGb,
				updateInstance,
				updateStatus,
			),
			composed.ComposeActions(
				"delete",
				removeReadyCondition,
				deleteInstance,
				waitInstanceDeleted,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}
