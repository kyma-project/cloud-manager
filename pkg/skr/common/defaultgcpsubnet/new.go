package defaultgcpsubnet

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// New returns a composed.Action that implements the following flow:
// * load GcpSubnet specified in the reconciled object
// * if no GcpSubnet found, find default SKR GcpSubnet
// * if no GcpSubnet found, create default SKR GcpSubnet with empty cidr
// * requeue until found or created GcpSubnet has Ready condition
// The provided state MUST implement State interface and object in the state
// MUST implement ObjWithGcpSubnetRef.
// The flow following this action can relay that the state will have non nil SKR GcpSubnet
// instance with Ready condition, unless object is marked for deletion and
// default GcpSubnet has to be created - then the state's SKR GcpSubnet will be nil. In other words,
// this action will not create default SKR GcpSubnet if object is marked for deletion.
func New() composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state, ok := st.(State)
		if !ok {
			return composed.LogErrorAndReturn(
				fmt.Errorf("state %T provided to defaultgcpsubnet flow does not implement defaultgcpsubnet.State", st),
				"Logical error",
				composed.StopAndForget,
				ctx,
			)
		}

		_, ok = state.Obj().(ObjWithGcpSubnetRef)
		if !ok {
			return composed.LogErrorAndReturn(
				fmt.Errorf("object %T provided to defaultgcpsubnet flow does not implement defaultgcpsubnet.ObjWithGcpSubnetRef", state.Obj()),
				"Logical error",
				composed.StopAndForget,
				ctx,
			)
		}

		return composed.ComposeActions(
			"defaultgcpsubnet",
			loadSpecifiedGcpSubnet,
			findDefaultSkrGcpSubnet,
			composed.If(composed.Not(composed.MarkedForDeletionPredicate),
				createDefaultSkrGcpSubnet,
				waitGcpSubnetReady,
			),
		)(ctx, state)
	}
}
