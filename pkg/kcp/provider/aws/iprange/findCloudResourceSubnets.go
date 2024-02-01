package iprange

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
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

	state.ObjAsIpRange().Status.Subnets = pie.Map(cloudResourcesSubnets, func(subnet ec2Types.Subnet) cloudcontrolv1beta1.IpRangeSubnet {
		return cloudcontrolv1beta1.IpRangeSubnet{
			Id:    pointer.StringDeref(subnet.SubnetId, ""),
			Zone:  pointer.StringDeref(subnet.AvailabilityZone, ""),
			Range: pointer.StringDeref(subnet.CidrBlock, ""),
		}
	})

	return nil, nil
}
