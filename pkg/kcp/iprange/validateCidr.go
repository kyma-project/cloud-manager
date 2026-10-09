package iprange

import (
	"context"
	"fmt"
	"net"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var reservedIPv4CIDRs = []string{
	"0.0.0.0/0",
	"127.0.0.0/8",
	"169.254.0.0/16",
}

func isReservedCIDR(input string) bool {
	_, inputNet, err := net.ParseCIDR(input)
	if err != nil {
		return false
	}
	for _, reserved := range reservedIPv4CIDRs {
		_, reservedNet, err := net.ParseCIDR(reserved)
		if err != nil {
			continue
		}
		if inputNet.String() == reservedNet.String() {
			return true
		}
	}
	return false
}

func validateCidr(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	cidr := state.ObjAsIpRange().Spec.Cidr
	if len(cidr) == 0 {
		return nil, ctx
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonInvalidCidr,
				Message: fmt.Sprintf("CIDR %s has invalid syntax", cidr),
			}).
			ErrorLogMessage("Error patching KCP IpRange status with invalid CIDR syntax").
			SuccessLogMsg("Forgetting KCP IpRange with invalid CIDR syntax").
			Run(ctx, state)
	}

	_, bits := ipNet.Mask.Size()
	if bits != 32 {
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonInvalidCidr,
				Message: fmt.Sprintf("CIDR %s is not IPv4", cidr),
			}).
			ErrorLogMessage("Error patching KCP IpRange status with non-IPv4 CIDR").
			SuccessLogMsg("Forgetting KCP IpRange with non-IPv4 CIDR").
			Run(ctx, state)
	}

	if isReservedCIDR(cidr) {
		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonInvalidCidr,
				Message: fmt.Sprintf("CIDR %s is a reserved range and cannot be used", cidr),
			}).
			ErrorLogMessage("Error patching KCP IpRange status with reserved CIDR").
			SuccessLogMsg("Forgetting KCP IpRange with reserved CIDR").
			Run(ctx, state)
	}

	return nil, ctx
}
