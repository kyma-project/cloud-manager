package dsl

import (
	"context"
	"errors"
	"fmt"
	"github.com/3th1nk/cidr"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"k8s.io/utils/ptr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsIpRangeWithSubnets(ctx context.Context, kcpClient ctrlclient.Client, client awsiprangeclient.Client, iprange *cloudcontrolv1beta1.IpRange,
	vpcId string, name string, iprangeCidr string,
) error {
	if iprange == nil {
		return errors.New("iprange given to CreateAwsIpRangeWithSubnets() can not be nil")
	}

	err := CreateKcpIpRange(ctx, kcpClient, iprange,
		WithName(name),
		WithKcpIpRangeRemoteRef(name),
		WithScope(name),
		WithKcpIpRangeSpecCidr(iprangeCidr),
	)
	if err != nil {
		return err
	}

	_, err = client.AssociateVpcCidrBlock(ctx, vpcId, iprangeCidr)
	if err != nil {
		return err
	}

	wholeRange, err := cidr.Parse(iprange.Spec.Cidr)
	if err != nil {
		return err
	}
	subnetRanges, err := wholeRange.SubNetting(cidr.MethodSubnetNum, 4)
	if err != nil {
		return err
	}
	subnetRanges = subnetRanges[:3]
	iprange.Status.Ranges = pie.Map(subnetRanges, func(c *cidr.CIDR) string {
		return c.CIDR().String()
	})

	zones := []string{"eu-west-1a", "eu-west-1b", "eu-west-1c"}
	for x, zone := range zones {
		rng := iprange.Status.Ranges[x]
		subnet, err := client.CreateSubnet(ctx, vpcId, zone, rng, awsutil.Ec2Tags(
			"Name", fmt.Sprintf("%s-%d", iprange.Name, x),
			common.TagCloudManagerName, name,
			common.TagCloudManagerRemoteName, iprange.Spec.RemoteRef.String(),
			common.TagScope, name,
			"cloud-manager.kyma-project.io/iprange", "1",
		))
		if err != nil {
			return err
		}

		iprange.Status.Subnets = append(iprange.Status.Subnets, cloudcontrolv1beta1.IpRangeSubnet{
			Id:    ptr.Deref(subnet.SubnetId, ""),
			Zone:  zone,
			Range: rng,
		})
	}

	err = UpdateStatus(ctx, kcpClient, iprange)
	if err != nil {
		return err
	}

	return nil
}
