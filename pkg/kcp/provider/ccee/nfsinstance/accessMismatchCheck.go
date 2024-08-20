package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func accessMismatchCheck(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.accessRight == nil {
		return nil, nil
	}

	mismatch := false
	if state.accessRight.ShareID != state.share.ID {
		logger.
			WithValues(
				"cceeAccessRightShareId", state.accessRight.ShareID,
				"cceeShareId", state.share.ID,
			).
			Info("CCEE NfsInstance access right share ID mismatch")
		mismatch = true
	}
	if state.accessRight.AccessType != "ip" {
		logger.
			WithValues("cceeAccessRightAccessType", state.accessRight.AccessType).
			Info("CCEE NfsInstance access right AccessType mismatch")
		mismatch = true
	}
	if state.accessRight.AccessTo != state.Scope().Spec.Scope.OpenStack.Network.Nodes {
		logger.
			WithValues(
				"cceeAccessRightAccessTo", state.accessRight.AccessTo,
				"cceeExpectedAccessTo", state.Scope().Spec.Scope.OpenStack.Network.Nodes,
			).
			Info("CCEE NfsInstance access right AccessTo mismatch")
		mismatch = true
	}
	if state.accessRight.AccessLevel != "rw" {
		logger.
			WithValues("cceeAccessRightAccessLevel", state.accessRight.AccessLevel).
			Info("CCEE NfsInstance access right AccessLevel mismatch")
		mismatch = true
	}

	if !mismatch {
		return nil, nil
	}

	// delete access right

	err := state.cceeClient.RevokeShareAccess(ctx, state.share.ID, state.accessRight.ID)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error revoking mismatched access right", composed.StopWithRequeue, ctx)
	}

	state.accessRight = nil

	logger.Info("CCEE NfsInstance mismatched access right was revoked")

	accessRightId, _ := state.ObjAsNfsInstance().GetStateData(StateDataAccessRightId)
	if accessRightId != "" {
		state.ObjAsNfsInstance().SetStateData(StateDataAccessRightId, "")

		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Failed patching CCEE NfsInstance status with revoked accessRightID").
			Run(ctx, state)
	}

	return nil, nil
}
