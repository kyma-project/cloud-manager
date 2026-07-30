package rediscluster

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloud "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// setSecurityIps ensures the default security IP group on the AliCloud Redis
// cluster contains the shoot VPC CIDR so that pods can reach the cluster.
// AliCloud creates instances with security IPs set to 127.0.0.1, blocking all
// inbound connections from the Kubernetes cluster.
// The action reads the current group before writing (drift check) and skips
// ModifySecurityIps when all required CIDRs are already present.
func setSecurityIps(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.instance == nil {
		return nil, ctx
	}

	scope := state.Scope()
	if scope.Spec.Scope.Alicloud == nil {
		return nil, ctx
	}

	nodesCidr := scope.Spec.Scope.Alicloud.Network.Nodes
	ipRangeCidr := ""
	if state.IpRange() != nil {
		ipRangeCidr = state.IpRange().Spec.Cidr
	}

	required := alicloud.BuildRequiredCidrs(nodesCidr, ipRangeCidr)
	if len(required) == 0 {
		return nil, ctx
	}

	existing, err := state.client.DescribeSecurityIps(ctx, state.instance.InstanceId)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error describing AliCloud r-kvstore cluster security IPs",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if alicloud.HasAllCidrs(existing, required) {
		return nil, ctx
	}

	if err := state.client.ModifySecurityIps(ctx, state.instance.InstanceId, strings.Join(required, ",")); err != nil {
		return composed.LogErrorAndReturn(err,
			"Error setting AliCloud r-kvstore cluster security IPs",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return nil, ctx
}
