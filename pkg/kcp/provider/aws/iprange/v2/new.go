package v2

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

// New returns an Action that will provision and deprovision resource in the cloud.
// Common post actions are executed after it in the common iprange flow
// so in the case of success it must return nil error as a signal of success.
// If it returns non-nil error then it will break the common iprange flow
// immediately so it must as well set the error conditions properly.
func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		ipRangeState := st.(iprangetypes.State)
		state, err := stateFactory.NewState(ctx, ipRangeState, logger)
		if err != nil {
			err = fmt.Errorf("error creating new aws iprange state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"awsIpRangeI2-main",
			vpcLoad,
			vpcFind,
			subnetsLoadAll,
			subnetsFindCloudResources,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"kcpIpRangeI2-create",
					preventCidrEdit,
					copyCidrToStatus,
					rangeSplitByZones,
					ensureShootZonesAndRangeSubnetsMatch,
					rangeCheckOverlap,
					rangeCheckBlockStatus,
					rangeCheckSubnetOverlap,
					rangeExtendVpcAddressSpace,
					subnetsCreate,
					subnetsCheckState,
					statusSuccess,
				),
				composed.ComposeActions(
					"kcpIpRangeI2-delete",
					statusRemoveReadyCondition,
					subnetsDelete,
					subnetsWaitDeleted,
					rangeDisassociateVpcAddressSpace,
					rangeWaitCidrBlockDisassociated,
				),
			),
		)(awsmeta.SetAwsAccountId(ctx, ipRangeState.Scope().Spec.Scope.Aws.AccountId), state)
	}
}
