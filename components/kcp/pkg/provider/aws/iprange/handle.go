package iprange

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/utils/pointer"
	"strings"
)

func handle(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	_ = composed.LoggerFromCtx(ctx)

	// TODO: api error UnauthorizedOperation: You are not authorized to perform this operation. User: arn:aws:iam::642531956841:user/cloudresource
	vpcNetworkName := state.Scope().Spec.Scope.Aws.VpcNetwork
	vpcList, err := state.networkClient.DescribeVpcs(ctx)
	if err != nil {
		return err, nil
	}

	var vpc *ec2Types.Vpc
loop:
	for _, v := range vpcList {
		if nameEquals(v.Tags, vpcNetworkName) {
			vpc = &v
			break loop
		}
	}
	if vpc == nil {
		return fmt.Errorf("vpc network %s not found", vpcNetworkName), nil
	}

	subnetList, err := state.networkClient.DescribeSubnets(ctx, pointer.StringDeref(vpc.VpcId, ""))
	if err != nil {
		return err, nil
	}

	var nodesSubnets []ec2Types.Subnet
	nodesSubnetPrefix := fmt.Sprintf("%s-nodes-z", vpcNetworkName)
	for _, sub := range subnetList {
		name := getNameFromTags(sub.Tags)
		if strings.HasPrefix(name, nodesSubnetPrefix) {
			nodesSubnets = append(nodesSubnets, sub)
		}
	}

	return nil, nil
}

func getNameFromTags(tags []ec2Types.Tag) string {
	for _, t := range tags {
		if pointer.StringDeref(t.Key, "") == "Name" {
			return pointer.StringDeref(t.Value, "")
		}
	}
	return ""
}

func nameEquals(tags []ec2Types.Tag, name string) bool {
	val := getNameFromTags(tags)
	return val == name
}
