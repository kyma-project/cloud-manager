package util

import ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

// Determinates whether VpcPeeringConnection can be deleted based on VPC peering connection lifecycle
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
