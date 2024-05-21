package v1

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func waitCidrBlockDisassociated(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	var theBlock *ec2Types.VpcCidrBlockAssociation
	for _, cidrBlock := range state.vpc.CidrBlockAssociationSet {
		if pointer.StringDeref(cidrBlock.CidrBlock, "") == state.ObjAsIpRange().Spec.Cidr {
			theBlock = &cidrBlock
		}
	}

	if theBlock == nil || theBlock.CidrBlockState == nil {
		return nil, nil
	}

	actMap := util.NewDelayActIgnoreBuilder[ec2Types.VpcCidrBlockStateCode](util.Ignore).
		Delay(ec2Types.VpcCidrBlockStateCodeDisassociating).
		Error(
			ec2Types.VpcCidrBlockStateCodeFailing,
			ec2Types.VpcCidrBlockStateCodeFailed,
			ec2Types.VpcCidrBlockStateCodeAssociated,
			ec2Types.VpcCidrBlockStateCodeAssociating,
		).
		Build()

	outcome := actMap.Case(theBlock.CidrBlockState.State)

	if outcome == util.Delay {
		logger.Info("Waiting for VPC Cidr block to get disassociated")
		return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
	}

	if outcome == util.Ignore {
		// all fine, it's disassociated
		return nil, nil
	}

	// it's in the failing/failed state, report the error, and stop and forget

	return composed.UpdateStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonUnknown,
			Message: "VPC Cidr block is in Failed/Associated state",
		}).
		SuccessError(composed.StopAndForget).
		Run(ctx, st)
}
