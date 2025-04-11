package v2

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

func rangeCheckBlockStatus(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	rangeCidr, _ := cidr.Parse(state.ObjAsIpRange().Status.Cidr)
	for _, set := range state.vpc.CidrBlockAssociationSet {
		cdr, err := cidr.Parse(ptr.Deref(set.CidrBlock, ""))
		if err != nil {
			logger.Error(err, "Error parsing AWS CIDR")
			continue
		}

		if util.CidrEquals(rangeCidr.CIDR(), cdr.CIDR()) &&
			// we must ignore disassociated sets
			!pie.Contains([]ec2types.VpcCidrBlockStateCode{
				ec2types.VpcCidrBlockStateCodeDisassociated,
				ec2types.VpcCidrBlockStateCodeDisassociating,
			}, set.CidrBlockState.State) {
			state.associatedCidrBlock = &set
			break
		}
	}

	if state.associatedCidrBlock == nil {
		logger.Info("Matching AWS CIDR block not found")
		return nil, nil
	}

	logger.
		WithValues(
			"cidrBlockAssociationId", state.associatedCidrBlock.AssociationId,
			"cidrBlockAssociationState", state.associatedCidrBlock.CidrBlockState.State,
			"cidrBlockAssociationMessage", state.associatedCidrBlock.CidrBlockState.StatusMessage,
		).
		Info("Found matching AWS CIDR block")

	if pie.Contains([]ec2types.VpcCidrBlockStateCode{
		ec2types.VpcCidrBlockStateCodeAssociated,
		ec2types.VpcCidrBlockStateCodeDisassociated,
		ec2types.VpcCidrBlockStateCodeDisassociating,
	}, state.associatedCidrBlock.CidrBlockState.State) {
		return nil, nil
	}
	if state.associatedCidrBlock.CidrBlockState.State == ec2types.VpcCidrBlockStateCodeAssociating {
		return composed.StopWithRequeueDelay(10 * time.Second), nil
	}

	return composed.PatchStatus(state.ObjAsIpRange()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonCidrAssociationFailed,
			Message: fmt.Sprintf("CIDR block status state is %s", state.associatedCidrBlock.CidrBlockState.State),
		}).
		ErrorLogMessage("Failed patching KCP IpRange CidrAssociationFailed status").
		SuccessLogMsg("Forgetting KCP IpRange with unhandled CidrBlock state").
		Run(ctx, st)
}
