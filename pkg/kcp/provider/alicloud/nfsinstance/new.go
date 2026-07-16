package nfsinstance

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
)

// New returns an Action that provisions and deprovisions an AliCloud NAS file system
// (with a VPC mount target and permission group) for a KCP NfsInstance.
func New(sf StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		nfsInstanceState := st.(nfsinstancetypes.State)

		if nfsInstanceState.Scope() == nil || nfsInstanceState.Scope().Spec.Scope.Alicloud == nil {
			if composed.MarkedForDeletionPredicate(ctx, st) {
				return nil, ctx
			}
			return composed.LogErrorAndReturn(
				errors.New("logical error"),
				"AliCloud NfsInstance flow called w/out alicloud scope",
				composed.StopAndForget,
				ctx,
			)
		}

		state, err := sf.NewState(ctx, nfsInstanceState)
		if err != nil {
			return fmt.Errorf("error creating alicloud nfsinstance state: %w", err), ctx
		}

		return composed.ComposeActionsNoName(
			loadFileSystem,
			loadAccessGroup,
			loadMountTargets,

			// create/update =========================================================================
			composed.If(
				composed.NotMarkedForDeletionPredicate,
				validateIpRangeSubnets,
				addFinalizer,
				createAccessGroup,
				createAccessRule,
				createFileSystem,
				waitFileSystemAvailable,
				createMountTargets,
				waitMountTargetsAvailable,
				updateStatus,
			),

			// delete ================================================================================
			composed.If(
				composed.MarkedForDeletionPredicate,
				removeReadyCondition,
				deleteMountTargets,
				waitMountTargetsDeleted,
				deleteFileSystem,
				waitFileSystemDeleted,
				deleteAccessGroup,
				removeFinalizer,
			),
		)(ctx, state)
	}
}
