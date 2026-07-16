package nfsinstance

import (
	"context"
	"slices"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// createAccessRule ensures NAS access rules exist for the cluster CIDRs (VPC and pods),
// granting read/write NFS access. Existing rules are not recreated.
func createAccessRule(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	desired := desiredAccessCidrs(state)
	if len(desired) == 0 {
		return nil, ctx
	}

	existing, err := state.client.DescribeAccessRules(ctx, state.accessGroupName)
	if err != nil {
		logger.Error(err, "Error loading AliCloud NAS access rules")
		return composed.StopWithRequeue, ctx
	}

	created := false
	for _, cidr := range desired {
		if slices.Contains(existing, cidr) {
			continue
		}
		logger.Info("Creating AliCloud NAS access rule", "accessGroupName", state.accessGroupName, "sourceCidr", cidr)
		if err := state.client.CreateAccessRule(ctx, state.accessGroupName, cidr); err != nil {
			logger.Error(err, "Error creating AliCloud NAS access rule", "sourceCidr", cidr)
			return composed.StopWithRequeue, ctx
		}
		created = true
	}

	if created {
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}

// desiredAccessCidrs returns the set of source CIDRs that should be allowed NFS access.
func desiredAccessCidrs(state *State) []string {
	net := state.Scope().Spec.Scope.Alicloud.Network
	var cidrs []string
	for _, c := range []string{net.VPC.CIDR, net.Pods, net.Nodes} {
		if c != "" && !slices.Contains(cidrs, c) {
			cidrs = append(cidrs, c)
		}
	}
	return cidrs
}
