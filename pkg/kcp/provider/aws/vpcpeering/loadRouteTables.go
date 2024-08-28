package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadRouteTables(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	routeTables, err := state.client.DescribeRouteTables(ctx, ptr.Deref(state.vpc.VpcId, "xxx"))

	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error loading AWS route tables", composed.LoggerIntoCtx(ctx, logger.WithValues("vpcId", ptr.Deref(state.vpc.VpcId, "xxx"))))
	}

	state.routeTables = routeTables

	return nil, nil
}
