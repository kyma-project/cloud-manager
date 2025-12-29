package state

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// RunStateMachine executes the state machine logic.
// TODO: Implement this action.
func RunStateMachine(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Implementation placeholder
	// This will:
	// 1. Determine current GCP state (from cached instance or operation)
	// 2. Map to CRD state
	// 3. Determine what operation is needed (ADD/MODIFY/DELETE/NONE)
	// 4. Set operation type in state
	// 5. Return nil, ctx to continue
	return nil, ctx
}
