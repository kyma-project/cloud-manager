package iprange

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
)

func findCloudResourceSubnets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	var cloudResourcesSubnets []ec2Types.Subnet
	for _, sub := range state.allSubnets {
		val := getTagValue(sub.Tags, tagKey)
		if len(val) > 0 {
			cloudResourcesSubnets = append(cloudResourcesSubnets, sub)
		}
	}

	state.cloudResourceSubnets = cloudResourcesSubnets

	return nil, nil
}
