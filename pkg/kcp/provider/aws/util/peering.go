package util

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

// IsTerminated identifies whether VpcPeeringConnection can be deleted based on VPC peering connection lifecycle
// https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-basics.html
func IsTerminated(peering *ec2types.VpcPeeringConnection) bool {
	code := peering.Status.Code
	if code == ec2types.VpcPeeringConnectionStateReasonCodeFailed ||
		code == ec2types.VpcPeeringConnectionStateReasonCodeExpired ||
		code == ec2types.VpcPeeringConnectionStateReasonCodeRejected ||
		code == ec2types.VpcPeeringConnectionStateReasonCodeDeleted {
		return true
	}

	return false
}

func ShouldUpdateRouteTable(tags []ec2types.Tag, mode cloudcontrolv1beta1.AwsRouteTableUpdateStrategy, tag string) bool {
	return mode == cloudcontrolv1beta1.AwsRouteTableUpdateStrategyAuto ||
		mode == cloudcontrolv1beta1.AwsRouteTableUpdateStrategyMatched && HasEc2Tag(tags, tag) ||
		mode == cloudcontrolv1beta1.AwsRouteTableUpdateStrategyUnmatched && !HasEc2Tag(tags, tag)
}

func IsRouteTableUpdateStrategyNone(mode cloudcontrolv1beta1.AwsRouteTableUpdateStrategy) bool {
	return mode == cloudcontrolv1beta1.AwsRouteTableUpdateStrategyNone
}
