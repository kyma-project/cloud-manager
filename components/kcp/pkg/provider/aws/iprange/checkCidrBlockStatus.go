package iprange

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/util"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func checkCidrBlockStatus(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	rangeCidr, _ := cidr.Parse(state.ObjAsIpRange().Spec.Cidr)
	for _, set := range state.vpc.CidrBlockAssociationSet {
		cdr, err := cidr.Parse(pointer.StringDeref(set.CidrBlock, ""))
		if err != nil {
			logger.Error(err, "Error parsing AWS CIDR: %w", err)
			continue
		}

		if util.CidrEquals(rangeCidr.CIDR(), cdr.CIDR()) {
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

	if state.associatedCidrBlock.CidrBlockState.State == ec2Types.VpcCidrBlockStateCodeAssociated {
		return nil, nil
	}
	if state.associatedCidrBlock.CidrBlockState.State == ec2Types.VpcCidrBlockStateCodeAssociating {
		return composed.StopWithRequeueDelay(10 * time.Second), nil
	}

	meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeError,
		Status:  "True",
		Reason:  cloudresourcesv1beta1.ReasonCidrAssociationFailed,
		Message: fmt.Sprintf("CIDR block status state is %s", state.associatedCidrBlock.CidrBlockState.State),
	})
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed updating CidrAssociationFailed status", composed.StopWithRequeue, nil)
	}

	return composed.StopAndForget, nil
}
