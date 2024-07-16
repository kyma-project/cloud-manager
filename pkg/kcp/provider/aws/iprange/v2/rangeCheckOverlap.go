package v2

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func rangeCheckOverlap(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	rangeCidr, _ := cidr.Parse(state.ObjAsIpRange().Status.Cidr)
	for _, set := range state.vpc.CidrBlockAssociationSet {
		cdr, err := cidr.Parse(ptr.Deref(set.CidrBlock, ""))
		if err != nil {
			logger.Error(err, "Error parsing AWS CIDR")
			continue
		}

		if util.CidrEquals(rangeCidr.CIDR(), cdr.CIDR()) {
			continue
		}

		if util.CidrOverlap(rangeCidr.CIDR(), cdr.CIDR()) {
			state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.ErrorState
			return composed.PatchStatus(state.ObjAsIpRange()).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudcontrolv1beta1.ConditionTypeError,
					Status:  "True",
					Reason:  cloudcontrolv1beta1.ReasonCidrOverlap,
					Message: fmt.Sprintf("CIDR overlaps with VPC adress range cidr %s", ptr.Deref(set.CidrBlock, "")),
				}).
				ErrorLogMessage("Error patching KCP IpRange status due to cidr overlap").
				SuccessLogMsg("Forgetting KCP IpRange due to cidr overlap").
				Run(ctx, st)
		}
	}

	return nil, nil
}
