package defaultiprange

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// New returns a composed.Action that implements the following flow:
// * load IpRange specified in the reconciled object
// * if no IpRange found, find default SKR IpRange
// * if no IpRange found, create default SKR IpRange with empty cidr
// * requeue until found or created IpRange has Ready condition
// The provided state MUST implement State interface and object in the state
// MUST implement ObjWithIpRangeRef.
// The flow following this action can relay that the state will have non nil SKR IpRange
// instance with Ready condition, unless object is marked for deletion and
// default IpRange has to be created - then the state's SKR IpRange will be nil. In other words,
// this action will not create default SKR IpRange if object is marked for deletion.
func New() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, ok := st.(State)
		if !ok {
			return composed.LogErrorAndReturn(
				fmt.Errorf("state %T provided to defaultiprange flow does not implement defaultiprange.State", st),
				"Logical error",
				composed.StopAndForget,
				ctx,
			)
		}

		_, ok = state.Obj().(ObjWithIpRangeRef)
		if !ok {
			return composed.LogErrorAndReturn(
				fmt.Errorf("object %T provided to defaultiprange flow does not implement defaultiprange.ObjWithIpRangeRef", state.Obj()),
				"Logical error",
				composed.StopAndForget,
				ctx,
			)
		}

		return composed.ComposeActions(
			"defaultiprange",
			checkIfIpRangeIsSpecified,
			loadSpecifiedIpRange,
			findDefaultSkrIpRange,
			createDefaultSkrIpRange,
			waitIpRangeReady,
		)(ctx, state)
	}
}
