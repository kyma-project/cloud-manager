package dsl

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/common"
	awsmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/mock"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"k8s.io/utils/ptr"
)

type AwsGardenerInfra struct {
	VPC         *ec2types.Vpc
	Subnets     []*ec2types.Subnet
	NatGateways []*ec2types.NatGateway
}

func CreateAwsGardenerResources(
	ctx context.Context,
	awsMock awsmock.AccountRegion,
	shootNamespace, shootName string,
	vpcCidr, nodesCidr string,
) (*AwsGardenerInfra, error) {
	result := &AwsGardenerInfra{}

	vpcName := common.GardenerVpcName(shootNamespace, shootName)

	wholeRange, err := cidr.Parse(nodesCidr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CIDR '%s': %v", vpcCidr, err)
	}

	numberOfSubnets := 4
	subnetRanges, err := wholeRange.SubNetting(cidr.MethodSubnetNum, numberOfSubnets)
	if err != nil {
		return nil, fmt.Errorf("error splitting vnet cidr %s: %w", vpcCidr, err)
	}

	var subnets []awsmock.VpcSubnet

	zones := []string{"a", "b", "c", "d"}
	for i := 1; i <= 3; i++ {
		sbnt := awsmock.VpcSubnet{
			AZ:   awsMock.Region() + zones[i-1],
			Cidr: subnetRanges[i-1].CIDR().String(),
			Tags: awsutil.Ec2Tags("Name", fmt.Sprintf("%s-%d", vpcName, i)),
		}
		subnets = append(subnets, sbnt)
	}

	vpcId := uuid.NewString()

	result.VPC = awsMock.AddVpc(vpcId, vpcCidr, awsutil.Ec2Tags("Name", vpcName), subnets)

	subnetList, err := awsMock.DescribeSubnets(ctx, vpcId)
	if err != nil {
		return nil, err
	}
	result.Subnets = pie.Map(subnetList, func(x ec2types.Subnet) *ec2types.Subnet {
		return ptr.To(x)
	})

	for _, subnet := range result.Subnets {
		gw, err := awsMock.AddNatGateway(vpcId, ptr.Deref(subnet.SubnetId, ""))
		if err != nil {
			return nil, err
		}
		result.NatGateways = append(result.NatGateways, gw)
	}

	return result, nil
}
