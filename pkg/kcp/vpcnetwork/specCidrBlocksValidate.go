package vpcnetwork

import (
	"context"
	"fmt"
	"slices"

	"github.com/3th1nk/cidr"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func specCidrBlocksValidate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	invalid, overlapping, normalized := specCidrBlocksValidateImpl(state.ObjAsVpcNetwork().Spec.CidrBlocks)

	if len(invalid) > 0 {
		state.ObjAsVpcNetwork().Status.State = string(cloudcontrolv1beta1.StateWarning)

		return composed.PatchStatus(state.ObjAsVpcNetwork()).
			SetExclusiveConditions(metav1.Condition{
				Type:               cloudcontrolv1beta1.ConditionTypeReady,
				Status:             metav1.ConditionFalse,
				ObservedGeneration: state.ObjAsVpcNetwork().Generation,
				Reason:             cloudcontrolv1beta1.ReasonInvalidCidr,
				Message:            fmt.Sprintf("Invalid CIDR blocks %v", invalid),
			}).
			Run(ctx, state)
	}

	if len(overlapping) > 0 {
		state.ObjAsVpcNetwork().Status.State = string(cloudcontrolv1beta1.StateWarning)

		return composed.PatchStatus(state.ObjAsVpcNetwork()).
			SetExclusiveConditions(
				metav1.Condition{
					Type:               cloudcontrolv1beta1.ConditionTypeReady,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: state.ObjAsVpcNetwork().Generation,
					Reason:             cloudcontrolv1beta1.ReasonCidrOverlap,
					Message:            fmt.Sprintf("Overlapping CIDR blocks %v", overlapping),
				},
				metav1.Condition{
					Type:               cloudcontrolv1beta1.ConditionTypeError,
					Status:             metav1.ConditionTrue,
					ObservedGeneration: state.ObjAsVpcNetwork().Generation,
					Reason:             cloudcontrolv1beta1.ReasonCidrOverlap,
					Message:            fmt.Sprintf("Overlapping CIDR blocks %v", overlapping),
				},
			).
			Run(ctx, state)
	}

	state.normalizedSpecCidrs = normalized

	return nil, ctx

}

func specCidrBlocksValidateImpl(cidrBlocks []string) (invalid []string, overlapping []string, normalized []string) {
	parsedCidrs := make([]*cidr.CIDR, 0, len(cidrBlocks))
	for _, x := range cidrBlocks {
		c, err := cidr.Parse(x)
		if err != nil {
			invalid = append(invalid, x)
			continue
		}
		parsedCidrs = append(parsedCidrs, c)
	}

	if len(invalid) > 0 {
		return
	}

	slices.SortStableFunc(parsedCidrs, func(a, b *cidr.CIDR) int {
		return util.IpCompare(a.StartIP(), b.StartIP())
	})

	for i := 0; i < len(parsedCidrs); i++ {
		for j := i + 1; j < len(parsedCidrs); j++ {
			if util.CidrOverlap(parsedCidrs[i].CIDR(), parsedCidrs[j].CIDR()) {
				overlapping = append(overlapping, fmt.Sprintf("[%s %s]", parsedCidrs[i].String(), parsedCidrs[j].String()))
			}
		}
	}

	if len(overlapping) > 0 {
		return
	}

	normalized = pie.Map(parsedCidrs, func(c *cidr.CIDR) string {
		return c.String()
	})

	return
}
