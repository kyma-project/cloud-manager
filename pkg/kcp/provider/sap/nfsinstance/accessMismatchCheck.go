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
				"sapAccessRightShareId", state.accessRight.ShareID,
				"sapShareId", state.share.ID,
			).
			Info("SAP NfsInstance access right share ID mismatch")
		mismatch = true
	}
	if state.accessRight.AccessType != "ip" {
		logger.
			WithValues("sapAccessRightAccessType", state.accessRight.AccessType).
			Info("SAP NfsInstance access right AccessType mismatch")
		mismatch = true
	}
	if state.accessRight.AccessTo != state.Scope().Spec.Scope.OpenStack.Network.Nodes {
		logger.
			WithValues(
				"sapAccessRightAccessTo", state.accessRight.AccessTo,
				"sapExpectedAccessTo", state.Scope().Spec.Scope.OpenStack.Network.Nodes,
			).
			Info("SAP NfsInstance access right AccessTo mismatch")
		mismatch = true
	}
	if state.accessRight.AccessLevel != "rw" {
		logger.
			WithValues("sapAccessRightAccessLevel", state.accessRight.AccessLevel).
			Info("SAP NfsInstance access right AccessLevel mismatch")
		mismatch = true
	}

	if !mismatch {
		return nil, nil
	}

	// delete access right

	err := state.sapClient.RevokeShareAccess(ctx, state.share.ID, state.accessRight.ID)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error revoking mismatched access right", composed.StopWithRequeue, ctx)
	}

	state.accessRight = nil

	logger.Info("SAP NfsInstance mismatched access right was revoked")

	accessRightId, _ := state.ObjAsNfsInstance().GetStateData(StateDataAccessRightId)
	if accessRightId != "" {
		state.ObjAsNfsInstance().SetStateData(StateDataAccessRightId, "")

		return composed.PatchStatus(state.ObjAsNfsInstance()).
			ErrorLogMessage("Failed patching SAP NfsInstance status with revoked accessRightID").
			Run(ctx, state)
	}

	return nil, nil
}
