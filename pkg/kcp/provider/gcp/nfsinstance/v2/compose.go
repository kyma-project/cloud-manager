package v2

import (
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// composeActions creates the main action pipeline for GCP NfsInstance reconciliation.
//
// Action flow:
//  1. AddCommonFinalizer - Ensure finalizer for cleanup
//  2. pollOperation - Poll pending GCP operations (if any)
//  3. loadInstance - Fetch instance from GCP Filestore API
//  4. Branch on deletion marker:
//     a. Not marked for deletion:
//     - createInstance - Create if instance doesn't exist
//     - waitInstanceReady - Wait for instance to be ready
//     - updateInstance - Update if instance needs changes
//     - updateStatus - Update NfsInstance status
//     b. Marked for deletion:
//     - deleteInstance - Delete if instance exists
//     - waitInstanceDeleted - Wait for deletion to complete
//     - RemoveCommonFinalizer - Remove finalizer
func composeActions() composed.Action {
	return composed.ComposeActions(
		"gcpNfsInstanceV2",
		// Setup phase
		actions.AddCommonFinalizer(),

		// Operation polling phase (handles any pending operations from previous reconciliation)
		pollOperation,

		// Load phase
		loadInstance,

		// Branching logic based on deletion marker
		composed.IfElse(
			composed.Not(composed.MarkedForDeletionPredicate),
			// Create/Update flow
			composed.ComposeActions(
				"create-update",
				createInstance,    // Create if instance doesn't exist
				waitInstanceReady, // Wait for instance to be ready
				updateInstance,    // Update if changes needed
				updateStatus,      // Update status
			),
			// Delete flow
			composed.ComposeActions(
				"delete",
				deleteInstance,      // Delete if instance exists
				waitInstanceDeleted, // Wait for deletion
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}
