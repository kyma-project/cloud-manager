package v2

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	// Import sub-packages when they are implemented
	// "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/operations"
	// "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/validation"
	// "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/state"
)

// composeActions creates the main action pipeline for GCP NfsInstance reconciliation.
//
// Action flow:
//  1. validatePreflight - Pre-flight validations (capacity, tier, network, IpRange)
//  2. AddCommonFinalizer - Ensure finalizer for cleanup
//  3. pollOperation - Poll pending GCP operations (if any)
//  4. loadInstance - Fetch instance from GCP Filestore API
//  5. validatePostCreate - Post-creation validation (no scale-down for BASIC tiers)
//  6. stateMachine - State machine logic + determine operation type
//  7. Branch on deletion marker:
//     a. Not marked for deletion:
//     - syncInstance - Create or update instance based on operation type
//     - pollOperation - Wait for GCP operation to complete
//     - updateStatus - Update NfsInstance status
//     - StopAndForget - Complete reconciliation
//     b. Marked for deletion:
//     - deleteInstance - Delete filestore instance
//     - pollOperation - Wait for deletion to complete
//     - RemoveCommonFinalizer - Remove finalizer
//     - StopAndForget - Complete reconciliation
func composeActions() composed.Action {
	return composed.ComposeActions(
		"gcpNfsInstanceV2",
		// Validation phase
		validatePreflight, // TODO: Implement in validation/preflight.go

		// Setup phase
		actions.AddCommonFinalizer(),

		// Operation polling phase (handles any pending operations from previous reconciliation)
		pollOperation, // TODO: Implement in operations/operation.go

		// Load phase
		loadInstance, // TODO: Implement in operations/load.go

		// Post-load validation phase
		validatePostCreate, // TODO: Implement in validation/postcreate.go

		// State machine phase (determines what operation is needed)
		runStateMachine, // TODO: Implement in state/machine.go

		// Branching logic based on deletion marker
		composed.IfElse(
			composed.Not(composed.MarkedForDeletionPredicate),
			// Create/Update flow
			composed.ComposeActions(
				"create-update",
				syncInstance,  // TODO: Implement in operations/create.go and operations/update.go
				pollOperation, // Wait for operation to complete
				updateStatus,  // TODO: Implement in state/status.go
				composed.StopAndForgetAction,
			),
			// Delete flow
			composed.ComposeActions(
				"delete",
				deleteInstance, // TODO: Implement in operations/delete.go
				pollOperation,  // Wait for deletion to complete
				actions.RemoveCommonFinalizer(),
				composed.StopAndForgetAction,
			),
		),
	)
}

// Placeholder action implementations
// These will be replaced by proper implementations in their respective packages

func validatePreflight(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Move to validation/preflight.go
	return nil, ctx
}

func pollOperation(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Move to operations/operation.go
	return nil, ctx
}

func loadInstance(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Move to operations/load.go
	return nil, ctx
}

func validatePostCreate(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Move to validation/postcreate.go
	return nil, ctx
}

func runStateMachine(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Move to state/machine.go
	return nil, ctx
}

func syncInstance(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Implement logic to call either create or update based on operation type
	// Will delegate to operations/create.go or operations/update.go
	return nil, ctx
}

func deleteInstance(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Move to operations/delete.go
	return nil, ctx
}

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Move to state/status.go
	return nil, ctx
}
