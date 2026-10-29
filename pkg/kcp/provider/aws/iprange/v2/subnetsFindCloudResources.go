package v2

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
)

func subnetsFindCloudResources(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// Subnets in a VPC can belong to multiple IpRanges. They all carry the
	// tagKey tag, so filtering on tagKey alone would mix subnets from sibling
	// IpRanges into this IpRange's state. Scope to the current IpRange by
	// matching the TagCloudManagerName tag written at subnet creation time.
	name := state.Name().String()

	var cloudResourcesSubnets []ec2types.Subnet
	for _, sub := range state.allSubnets {
		if len(awsutil.GetEc2TagValue(sub.Tags, tagKey)) == 0 {
			continue
		}
		if awsutil.GetEc2TagValue(sub.Tags, common.TagCloudManagerName) != name {
			continue
		}
		cloudResourcesSubnets = append(cloudResourcesSubnets, sub)
	}

	state.cloudResourceSubnets = cloudResourcesSubnets

	return nil, nil
}
