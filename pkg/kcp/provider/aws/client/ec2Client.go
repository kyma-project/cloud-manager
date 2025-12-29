package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"k8s.io/utils/ptr"
)

type Ec2Client interface {
	DescribeNatGateway(ctx context.Context, vpcId string) ([]ec2types.NatGateway, error)

	CreateInternetGateway(ctx context.Context, name string) (*ec2types.InternetGateway, error)
	AttachInternetGateway(ctx context.Context, vpcId, internetGatewayId string) error
	DescribeInternetGateways(ctx context.Context, name string) ([]ec2types.InternetGateway, error)
	DeleteInternetGateway(ctx context.Context, internetGatewayId string) error

	DescribeDhcpOptions(ctx context.Context, name string) ([]ec2types.DhcpOptions, error)
	CreateDhcpOptions(ctx context.Context, name, domainName string, tags []ec2types.Tag) (*ec2types.DhcpOptions, error)
	AssociateDhcpOptions(ctx context.Context, vpcId string, dhcpOptionsId string) error
	DeleteDhcpOptions(ctx context.Context, dhcpOptionsId string) error

	CreateVpc(ctx context.Context, name, cidr string, tags []ec2types.Tag) (*ec2types.Vpc, error)
	DeleteVpc(ctx context.Context, vpcId string) error
	DescribeVpc(ctx context.Context, vpcId string) (*ec2types.Vpc, error)
	DescribeVpcs(ctx context.Context, name string) ([]ec2types.Vpc, error)
	AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*ec2types.VpcCidrBlockAssociation, error)
	DisassociateVpcCidrBlockInput(ctx context.Context, associationId string) error

	DescribeSubnets(ctx context.Context, vpcId string) ([]ec2types.Subnet, error)
	DescribeSubnet(ctx context.Context, subnetId string) (*ec2types.Subnet, error)
	CreateSubnet(ctx context.Context, vpcId, az, cidr string, tags []ec2types.Tag) (*ec2types.Subnet, error)
	DeleteSubnet(ctx context.Context, subnetId string) error

	DescribeSecurityGroups(ctx context.Context, filters []ec2types.Filter, groupIds []string) ([]ec2types.SecurityGroup, error)
	CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2types.Tag) (string, error)
	DeleteSecurityGroup(ctx context.Context, id string) error
	AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2types.IpPermission) error

	CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string, tag []ec2types.Tag) (*ec2types.VpcPeeringConnection, error)
	DescribeVpcPeeringConnection(ctx context.Context, vpcPeeringConnectionId string) (*ec2types.VpcPeeringConnection, error)
	DescribeVpcPeeringConnections(ctx context.Context) ([]ec2types.VpcPeeringConnection, error)
	AcceptVpcPeeringConnection(ctx context.Context, connectionId *string) (*ec2types.VpcPeeringConnection, error)
	DeleteVpcPeeringConnection(ctx context.Context, connectionId *string) error
	DescribeRouteTables(ctc context.Context, vpcId string) ([]ec2types.RouteTable, error)
	CreateRoute(ctx context.Context, routeTableId, destinationCidrBlock, vpcPeeringConnectionId *string) error
	DeleteRoute(ctx context.Context, routeTableId, destinationCidrBlock *string) error
}

func NewEc2Client(svc *ec2.Client) Ec2Client {
	return &ec2Client{
		svc: svc,
	}
}

var _ Ec2Client = (*ec2Client)(nil)

type ec2Client struct {
	svc *ec2.Client
}

func (c *ec2Client) CreateInternetGateway(ctx context.Context, name string) (*ec2types.InternetGateway, error) {
	in := &ec2.CreateInternetGatewayInput{
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInternetGateway,
				Tags: []ec2types.Tag{
					{
						Key:   ptr.To("Name"),
						Value: ptr.To(name),
					},
				},
			},
		},
	}
	out, err := c.svc.CreateInternetGateway(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.InternetGateway, nil
}

func (c *ec2Client) AttachInternetGateway(ctx context.Context, vpcId, internetGatewayId string) error {
	in := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: ptr.To(internetGatewayId),
		VpcId:             ptr.To(vpcId),
	}
	_, err := c.svc.AttachInternetGateway(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

func (c *ec2Client) DescribeInternetGateways(ctx context.Context, name string) ([]ec2types.InternetGateway, error) {
	in := &ec2.DescribeInternetGatewaysInput{}
	if name != "" {
		in.Filters = []ec2types.Filter{
			{
				Name:   ptr.To("tag:Name"),
				Values: []string{name},
			},
		}
	}
	out, err := c.svc.DescribeInternetGateways(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.InternetGateways, nil
}

func (c *ec2Client) DeleteInternetGateway(ctx context.Context, internetGatewayId string) error {
	_, err := c.svc.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: ptr.To(internetGatewayId),
	})
	return err
}

func (c *ec2Client) DescribeNatGateway(ctx context.Context, vpcId string) ([]ec2types.NatGateway, error) {
	in := &ec2.DescribeNatGatewaysInput{
		Filter: []ec2types.Filter{
			{
				Name:   ptr.To("vpc-id"),
				Values: []string{vpcId},
			},
		},
	}
	resp, err := c.svc.DescribeNatGateways(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp.NatGateways, nil
}

func (c *ec2Client) AssociateDhcpOptions(ctx context.Context, vpcId string, dhcpOptionsId string) error {
	in := &ec2.AssociateDhcpOptionsInput{
		VpcId:         &vpcId,
		DhcpOptionsId: &dhcpOptionsId,
	}
	_, err := c.svc.AssociateDhcpOptions(ctx, in)
	return err
}

func (c *ec2Client) DeleteDhcpOptions(ctx context.Context, dhcpOptionsId string) error {
	_, err := c.svc.DeleteDhcpOptions(ctx, &ec2.DeleteDhcpOptionsInput{
		DhcpOptionsId: ptr.To(dhcpOptionsId),
	})
	return err
}

func (c *ec2Client) CreateDhcpOptions(ctx context.Context, name, domainName string, tags []ec2types.Tag) (*ec2types.DhcpOptions, error) {
	tags = pie.FilterNot(tags, func(tag ec2types.Tag) bool {
		return ptr.Deref(tag.Key, "") == "Name"
	})
	tags = pie.Insert(tags, 0, ec2types.Tag{
		Key:   ptr.To("Name"),
		Value: ptr.To(name),
	})
	in := &ec2.CreateDhcpOptionsInput{
		DhcpConfigurations: []ec2types.NewDhcpConfiguration{
			{
				Key:    ptr.To("domain-name"),
				Values: []string{domainName},
			},
			{
				Key:    ptr.To("domain-name-servers"),
				Values: []string{"AmazonProvidedDNS"},
			},
		},
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeDhcpOptions,
				Tags:         tags,
			},
		},
	}
	out, err := c.svc.CreateDhcpOptions(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.DhcpOptions, nil
}

func (c *ec2Client) DescribeDhcpOptions(ctx context.Context, name string) ([]ec2types.DhcpOptions, error) {
	in := &ec2.DescribeDhcpOptionsInput{}
	if name != "" {
		in.Filters = []ec2types.Filter{
			{
				Name:   ptr.To("tag:Name"),
				Values: []string{name},
			},
		}
	}
	out, err := c.svc.DescribeDhcpOptions(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.DhcpOptions, nil
}

func (c *ec2Client) DeleteVpc(ctx context.Context, vpcId string) error {
	_, err := c.svc.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: ptr.To(vpcId),
	})
	return err
}

func (c *ec2Client) CreateVpc(ctx context.Context, name, cidr string, tags []ec2types.Tag) (*ec2types.Vpc, error) {
	tags = pie.FilterNot(tags, func(tag ec2types.Tag) bool {
		return ptr.Deref(tag.Key, "") == "Name"
	})
	tags = pie.Insert(tags, 0, ec2types.Tag{
		Key:   ptr.To("Name"),
		Value: ptr.To(name),
	})
	in := &ec2.CreateVpcInput{
		CidrBlock: ptr.To(cidr),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeVpc,
				Tags:         tags,
			},
		},
	}
	out, err := c.svc.CreateVpc(ctx, in)
	if err != nil {
		return nil, err
	}

	_, err = c.svc.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
		VpcId: out.Vpc.VpcId,
		EnableDnsSupport: &ec2types.AttributeBooleanValue{
			Value: ptr.To(true),
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = c.svc.ModifyVpcAttribute(ctx, &ec2.ModifyVpcAttributeInput{
		VpcId: out.Vpc.VpcId,
		EnableDnsHostnames: &ec2types.AttributeBooleanValue{
			Value: ptr.To(true),
		},
	})
	if err != nil {
		return nil, err
	}

	return out.Vpc, nil
}

func (c *ec2Client) DescribeVpc(ctx context.Context, vpcId string) (*ec2types.Vpc, error) {
	out, err := c.svc.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcId},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Vpcs) > 0 {
		return &out.Vpcs[0], nil
	}
	return nil, nil
}

func (c *ec2Client) DescribeVpcs(ctx context.Context, name string) ([]ec2types.Vpc, error) {
	in := &ec2.DescribeVpcsInput{}
	if name != "" {
		in.Filters = []ec2types.Filter{
			{
				Name:   ptr.To("tag:Name"),
				Values: []string{name},
			},
		}
	}
	out, err := c.svc.DescribeVpcs(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.Vpcs, nil
}

func (c *ec2Client) AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*ec2types.VpcCidrBlockAssociation, error) {
	in := &ec2.AssociateVpcCidrBlockInput{
		VpcId:     aws.String(vpcId),
		CidrBlock: aws.String(cidr),
	}
	out, err := c.svc.AssociateVpcCidrBlock(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.CidrBlockAssociation, nil
}

func (c *ec2Client) DisassociateVpcCidrBlockInput(ctx context.Context, associationId string) error {
	in := &ec2.DisassociateVpcCidrBlockInput{AssociationId: &associationId}
	_, err := c.svc.DisassociateVpcCidrBlock(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

func (c *ec2Client) DescribeSubnets(ctx context.Context, vpcId string) ([]ec2types.Subnet, error) {
	out, err := c.svc.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{
				Name:   ptr.To("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return out.Subnets, nil
}

func (c *ec2Client) CreateSubnet(ctx context.Context, vpcId, az, cidr string, tags []ec2types.Tag) (*ec2types.Subnet, error) {
	in := &ec2.CreateSubnetInput{
		VpcId:            ptr.To(vpcId),
		AvailabilityZone: ptr.To(az),
		CidrBlock:        ptr.To(cidr),
	}
	if len(tags) > 0 {
		in.TagSpecifications = []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSubnet,
				Tags:         tags,
			},
		}
	}
	out, err := c.svc.CreateSubnet(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.Subnet, nil
}

func (c *ec2Client) DeleteSubnet(ctx context.Context, subnetId string) error {
	_, err := c.svc.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: ptr.To(subnetId),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *ec2Client) DescribeSubnet(ctx context.Context, subnetId string) (*ec2types.Subnet, error) {
	out, err := c.svc.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{
				Name:   ptr.To("subnet-id"),
				Values: []string{subnetId},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Subnets) > 1 {
		return nil, fmt.Errorf("expected at most one subnet by id, but got: %v", pie.Map(out.Subnets, func(s ec2types.Subnet) string {
			return ptr.Deref(s.SubnetId, "")
		}))
	}
	var result *ec2types.Subnet
	if len(out.Subnets) > 0 {
		result = &out.Subnets[0]
	}
	return result, nil
}

func (c *ec2Client) DescribeSecurityGroups(ctx context.Context, filters []ec2types.Filter, groupIds []string) ([]ec2types.SecurityGroup, error) {
	out, err := c.svc.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters:  filters,
		GroupIds: groupIds,
	})
	if err != nil {
		return nil, err
	}
	return out.SecurityGroups, nil
}

func (c *ec2Client) CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2types.Tag) (string, error) {
	out, err := c.svc.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		Description: ptr.To(name),
		GroupName:   ptr.To(name),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSecurityGroup,
				Tags:         tags,
			},
		},
		VpcId: ptr.To(vpcId),
	})
	if err != nil {
		return "", err
	}
	return ptr.Deref(out.GroupId, ""), nil
}

func (c *ec2Client) DeleteSecurityGroup(ctx context.Context, id string) error {
	in := &ec2.DeleteSecurityGroupInput{
		GroupId: ptr.To(id),
	}
	_, err := c.svc.DeleteSecurityGroup(ctx, in)
	return err
}

func (c *ec2Client) AuthorizeSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2types.IpPermission) error {
	_, err := c.svc.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       ptr.To(groupId),
		IpPermissions: ipPermissions,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *ec2Client) CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string, tags []ec2types.Tag) (*ec2types.VpcPeeringConnection, error) {
	out, err := c.svc.CreateVpcPeeringConnection(ctx, &ec2.CreateVpcPeeringConnectionInput{
		VpcId:       vpcId,
		PeerVpcId:   remoteVpcId,
		PeerRegion:  remoteRegion,
		PeerOwnerId: remoteAccountId,
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeVpcPeeringConnection,
				Tags:         tags,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return out.VpcPeeringConnection, nil
}

func (c *ec2Client) DescribeVpcPeeringConnections(ctx context.Context) ([]ec2types.VpcPeeringConnection, error) {
	out, err := c.svc.DescribeVpcPeeringConnections(ctx, &ec2.DescribeVpcPeeringConnectionsInput{})
	if err != nil {
		return nil, err
	}
	return out.VpcPeeringConnections, err
}

func (c *ec2Client) DescribeVpcPeeringConnection(ctx context.Context, vpcPeeringConnectionId string) (*ec2types.VpcPeeringConnection, error) {
	out, err := c.svc.DescribeVpcPeeringConnections(ctx, &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: []string{vpcPeeringConnectionId},
	})
	if err != nil {
		return nil, err
	}

	if len(out.VpcPeeringConnections) > 0 {
		return &out.VpcPeeringConnections[0], nil
	}
	return nil, nil
}

func (c *ec2Client) AcceptVpcPeeringConnection(ctx context.Context, connectionId *string) (*ec2types.VpcPeeringConnection, error) {
	out, err := c.svc.AcceptVpcPeeringConnection(ctx, &ec2.AcceptVpcPeeringConnectionInput{
		VpcPeeringConnectionId: connectionId,
	})

	if err != nil {
		return nil, err
	}

	return out.VpcPeeringConnection, nil
}

func (c *ec2Client) DeleteVpcPeeringConnection(ctx context.Context, connectionId *string) error {
	_, err := c.svc.DeleteVpcPeeringConnection(ctx, &ec2.DeleteVpcPeeringConnectionInput{
		VpcPeeringConnectionId: connectionId,
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *ec2Client) DescribeRouteTables(ctx context.Context, vpcId string) ([]ec2types.RouteTable, error) {
	out, err := c.svc.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []ec2types.Filter{
			{
				Name:   ptr.To("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return out.RouteTables, nil
}

func (c *ec2Client) CreateRoute(ctx context.Context, routeTableId, destinationCidrBlock, vpcPeeringConnectionId *string) error {
	_, err := c.svc.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:           routeTableId,
		DestinationCidrBlock:   destinationCidrBlock,
		VpcPeeringConnectionId: vpcPeeringConnectionId,
	})

	return err
}

func (c *ec2Client) DeleteRoute(ctx context.Context, routeTableId, destinationCidrBlock *string) error {
	_, err := c.svc.DeleteRoute(ctx, &ec2.DeleteRouteInput{
		RouteTableId:         routeTableId,
		DestinationCidrBlock: destinationCidrBlock,
	})
	return err
}
