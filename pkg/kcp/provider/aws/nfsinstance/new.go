package nfsinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nfsinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		nfsState := st.(nfsinstancetypes.State)
		state, err := stateFactory.NewState(ctx, nfsState)
		if err != nil {
			err = fmt.Errorf("error creating new aws nfsinstance state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}
		return composed.ComposeActions(
			"awsNfsInstance",
			composed.BuildSwitchAction(
				"awsNfsInstance-switch",
				// non-delete
				composed.ComposeActions(
					"awsNfsInstance-non-delete",
					validateIpRangeSubnets,
					addFinalizer,
					findSecurityGroup,
					createSecurityGroup,
					loadSecurityGroup,
					authorizeSecurityGroupIngress,
					loadEfs,
					createEfs,
					waitEfsAvailable,
					loadMountTargets,
					validateExistingMountTargets,
					createMountTargets,
					waitMountTargetsAvailable,
					removeMountTargetsFromOtherVpcs,
					updateStatus,

					composed.StopAndForgetAction,
				),
				// delete
				composed.NewCase(
					composed.MarkedForDeletionPredicate,
					composed.ComposeActions(
						"awsNfsInstance-delete",
						removeReadyCondition,
						loadEfs,
						findSecurityGroup,
						loadMountTargets,

						deleteMountTargets,
						waitMountTargetsDeleted,

						deleteEfs,
						waitEfsDeleted,

						deleteSecurityGroup,

						removeFinalizer,

						composed.StopAndForgetAction,
					),
				),
			), // switch
			composed.StopAndForgetAction,
		)(awsmeta.SetAwsAccountId(ctx, nfsState.Scope().Spec.Scope.Aws.AccountId), state)
	}
}
