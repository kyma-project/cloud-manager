package vpcnetwork

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/3th1nk/cidr"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	awsvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcnetwork/client"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

type CreateInfraOption func(*createInfraOptions)

func WithName(name string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.name = name
	}
}

func WithCidrBlocks(cidrBlocks []string) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.cidrBlocks = append(o.cidrBlocks, cidrBlocks...)
	}
}

func WithClient(c awsvpcnetworkclient.Client) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.client = c
	}
}

func WithTimeout(t time.Duration) CreateInfraOption {
	return func(o *createInfraOptions) {
		o.timeout = t
	}
}

type createInfraOptions struct {
	name       string
	cidrBlocks []string
	client     awsvpcnetworkclient.Client
	timeout    time.Duration
}

type CreateInfraOutput struct {
	Created                bool
	Updated                bool
	Vpc                    *ec2types.Vpc
	DefaultSecurityGroupId string
	InternetGateway        *ec2types.InternetGateway
}

func (o *createInfraOptions) validate() error {
	var result error
	if o.name == "" {
		result = errors.Join(result, fmt.Errorf("name is required"))
	}
	if len(o.cidrBlocks) == 0 {
		result = errors.Join(result, fmt.Errorf("at least one cidr block is required"))
	}
	for _, c := range o.cidrBlocks {
		_, err := cidr.Parse(c)
		if err != nil {
			result = errors.Join(result, fmt.Errorf("invalid cidr block %q: %w", c, err))
		}
	}
	if o.client == nil {
		result = errors.Join(result, fmt.Errorf("client is required"))
	}
	if o.timeout == 0 {
		o.timeout = 5 * time.Minute
	}
	return result
}

func CreateInfra(ctx context.Context, opts ...CreateInfraOption) (*CreateInfraOutput, error) {
	created := false
	updated := false
	o := &createInfraOptions{}
	for _, opt := range opts {
		opt(o)
	}
	if err := o.validate(); err != nil {
		return nil, err
	}

	// load VPC
	vpcArr, err := o.client.DescribeVpcs(ctx, o.name)
	if err != nil {
		return nil, fmt.Errorf("error describing vpc: %w", err)
	}
	var vpc *ec2types.Vpc
	if len(vpcArr) == 0 {
		v, err := o.client.CreateVpc(ctx, o.name, o.cidrBlocks[0], nil)
		if err != nil {
			return nil, fmt.Errorf("error creating vpc: %w", err)
		}
		vpc = v
		created = true
	} else if len(vpcArr) == 1 {
		vpc = &vpcArr[0]
	} else {
		return nil, fmt.Errorf("multiple vpcs found with name %s", o.name)
	}

	// wait VPC available

	if vpc.State != ec2types.VpcStateAvailable {
		err = wait.PollUntilContextTimeout(ctx, time.Second, o.timeout, false, func(ctx context.Context) (done bool, err error) {
			v, err := o.client.DescribeVpc(ctx, ptr.Deref(vpc.VpcId, ""))
			if err != nil {
				return false, err
			}
			vpc = v
			return vpc.State == ec2types.VpcStateAvailable, nil
		})
		if err != nil {
			return nil, fmt.Errorf("error describing vpc while polling to become available: %w", err)
		}
	}

	// validate primary vpc block didn't change

	if ptr.Deref(vpc.CidrBlock, "") != o.cidrBlocks[0] {
		return nil, fmt.Errorf("primary cidr block can not change - %s was changed to %s", ptr.Deref(vpc.CidrBlock, ""), o.cidrBlocks[0])
	}

	// remove cidr blocks

	// assocID => cidrBlock
	removeCidrAssociations := map[string]string{}
	for _, assoc := range vpc.CidrBlockAssociationSet {
		if ptr.Deref(assoc.CidrBlock, "xxx") == ptr.Deref(vpc.CidrBlock, "yyy") {
			continue
		}
		found := false
		for _, c := range o.cidrBlocks {
			if c == ptr.Deref(assoc.CidrBlock, "") {
				found = true
				break
			}
		}
		if !found {
			removeCidrAssociations[ptr.Deref(assoc.AssociationId, "")] = ptr.Deref(assoc.CidrBlock, "")
		}
	}
	for assocId, cidrBlock := range removeCidrAssociations {
		err = o.client.DisassociateVpcCidrBlockInput(ctx, assocId)
		if err != nil {
			return nil, fmt.Errorf("error disassociating cidr block %q: %w", cidrBlock, err)
		}
	}
	if len(removeCidrAssociations) > 0 {
		err = waitVpcCidrBlocksAvailable(ctx, vpc, o)
		if err != nil {
			return nil, fmt.Errorf("after removing cidr blocks %v: %w", pie.Values(removeCidrAssociations), err)
		}
		updated = true
	}

	// add cidr blocks

	var addNewCidrs []string
	for _, c := range o.cidrBlocks[1:] {
		found := false
		for _, a := range vpc.CidrBlockAssociationSet {
			if c == ptr.Deref(a.CidrBlock, "") {
				found = true
				break
			}
		}
		if !found {
			addNewCidrs = append(addNewCidrs, c)
		}
	}
	for _, c := range addNewCidrs {
		_, err = o.client.AssociateVpcCidrBlock(ctx, ptr.Deref(vpc.VpcId, ""), c)
		if err != nil {
			return nil, fmt.Errorf("error associating cidr block to vpc: %w", err)
		}
	}
	if len(addNewCidrs) > 0 {
		err = waitVpcCidrBlocksAvailable(ctx, vpc, o)
		if err != nil {
			return nil, fmt.Errorf("after adding cidr blocks %v: %w", addNewCidrs, err)
		}
		updated = true
	}

	// dhcp options

	var do *ec2types.DhcpOptions
	doArr, err := o.client.DescribeDhcpOptions(ctx, o.name)
	if err != nil {
		return nil, fmt.Errorf("error describing dhcp options: %w", err)
	}
	if len(doArr) == 0 {
		domainName := "ec2.internal"
		if o.client.Region() != "us-east-1" {
			domainName = fmt.Sprintf("%s.compute.internal", o.client.Region())
		}
		dodo, err := o.client.CreateDhcpOptions(ctx, o.name, domainName, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating dhcp options: %w", err)
		}
		do = dodo
		updated = true
	} else if len(doArr) == 1 {
		do = &doArr[0]
	} else {
		return nil, fmt.Errorf("multiple dhcp options found with name %q", o.name)
	}

	if ptr.Deref(vpc.DhcpOptionsId, "xxx") != ptr.Deref(do.DhcpOptionsId, "yyy") {
		err = o.client.AssociateDhcpOptions(ctx, ptr.Deref(vpc.VpcId, ""), ptr.Deref(do.DhcpOptionsId, ""))
		if err != nil {
			return nil, fmt.Errorf("error associating dhcp options: %w", err)
		}
		updated = true
	}

	// default security group

	var sgID string
	sgArr, err := o.client.DescribeSecurityGroups(ctx, []ec2types.Filter{
		{
			Name:   ptr.To("vpc-id"),
			Values: []string{ptr.Deref(vpc.VpcId, "")},
		},
		{
			Name:   ptr.To("tag:Name"),
			Values: []string{o.name},
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error describing security groups: %w", err)
	}
	if len(sgArr) == 0 {
		id, err := o.client.CreateSecurityGroup(ctx, ptr.Deref(vpc.VpcId, ""), o.name, []ec2types.Tag{
			{
				Key:   ptr.To("Name"),
				Value: ptr.To(o.name),
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error creating security group: %w", err)
		}
		sgID = id
		updated = true
	} else if len(sgArr) > 1 {
		return nil, fmt.Errorf("multiple security groups found with name %q", o.name)
	}

	// internet gateway

	igwArr, err := o.client.DescribeInternetGateways(ctx, o.name)
	if err != nil {
		return nil, fmt.Errorf("error describing internet gateways: %w", err)
	}
	var igw *ec2types.InternetGateway
	if len(igwArr) == 1 {
		igw = &igwArr[0]
	} else if len(igwArr) == 0 {
		x, err := o.client.CreateInternetGateway(ctx, o.name)
		if err != nil {
			return nil, fmt.Errorf("error creating internet gateway: %w", err)
		}
		igw = x
		updated = true
	} else {
		return nil, fmt.Errorf("multiple internet gateways found with name %q", o.name)
	}

	isAttached := false
	for _, att := range igw.Attachments {
		if ptr.Deref(att.VpcId, "") == ptr.Deref(vpc.VpcId, "") {
			isAttached = true
			break
		}
	}
	if !isAttached {
		err = o.client.AttachInternetGateway(ctx, ptr.Deref(vpc.VpcId, ""), ptr.Deref(igw.InternetGatewayId, ""))
		if err != nil {
			return nil, fmt.Errorf("error attaching internet gateway: %w", err)
		}
		updated = true
	}

	return &CreateInfraOutput{
		Created:                created,
		Updated:                updated,
		Vpc:                    vpc,
		DefaultSecurityGroupId: sgID,
		InternetGateway:        igw,
	}, nil
}

func DeleteInfra(ctx context.Context, name string, c awsvpcnetworkclient.Client) error {
	/*
		* igw delete
		  * detach from vpc
		  * delete igw
		* sg delete
		* vpc delete
	*/

	// load vpc
	vpcArr, err := c.DescribeVpcs(ctx, name)
	if err != nil {
		return fmt.Errorf("error describing vpc: %w", err)
	}

	// assuming only igw with that name is attached to the vpc(s)
	// theoretically there could be other igws not attached to the vpc(s), but optimistically we ignore them

	// load internet gateways
	igwArr, err := c.DescribeInternetGateways(ctx, name)
	if err != nil {
		return fmt.Errorf("error describing internet gateways: %w", err)
	}

	for _, vpc := range vpcArr {

		for _, igw := range igwArr {
			// detach internet gateway from vpc
			for _, att := range igw.Attachments {
				if ptr.Deref(att.VpcId, "xxx") == ptr.Deref(vpc.VpcId, "yyy") {
					if err := c.DetachInternetGateway(ctx, ptr.Deref(vpc.VpcId, ""), ptr.Deref(igw.InternetGatewayId, "")); err != nil {
						return fmt.Errorf("error detaching internet gateway %q from vpc %q: %w", ptr.Deref(igw.InternetGatewayId, ""), ptr.Deref(vpc.VpcId, ""), err)
					}
				}
			}
		}

		// delete all security groups from vpc
		sgArr, err := c.DescribeSecurityGroups(ctx, []ec2types.Filter{
			{
				Name:   ptr.To("vpc-id"),
				Values: []string{ptr.Deref(vpc.VpcId, "")},
			},
			{
				Name:   ptr.To("tag:Name"),
				Values: []string{name},
			},
		}, nil)
		if err != nil {
			return fmt.Errorf("error describing security groups: %w", err)
		}
		for _, sg := range sgArr {
			if err := c.DeleteSecurityGroup(ctx, ptr.Deref(sg.GroupId, "")); err != nil {
				return fmt.Errorf("error deleting security group %q: %w", ptr.Deref(sg.GroupId, ""), err)
			}
		}

		if err := c.DeleteVpc(ctx, ptr.Deref(vpc.VpcId, "")); err != nil {
			return fmt.Errorf("error deleting vpc: %w", err)
		}
	} // for each vpc

	// delete internet gateway
	for _, igw := range igwArr {
		if err := c.DeleteInternetGateway(ctx, ptr.Deref(igw.InternetGatewayId, "")); err != nil {
			return fmt.Errorf("error deleting internet gateway: %w", err)
		}
	}

	// dhcp options

	doArr, err := c.DescribeDhcpOptions(ctx, name)
	if err != nil {
		return fmt.Errorf("error describing dhcp options: %w", err)
	}
	for _, do := range doArr {
		if err := c.DeleteDhcpOptions(ctx, ptr.Deref(do.DhcpOptionsId, "")); err != nil {
			return fmt.Errorf("error deleting dhcp options: %w", err)
		}
	}

	return nil
}

func waitVpcCidrBlocksAvailable(ctx context.Context, vpc *ec2types.Vpc, o *createInfraOptions) error {
	err := wait.PollUntilContextTimeout(ctx, time.Second, o.timeout, false, func(ctx context.Context) (done bool, err error) {
		v, err := o.client.DescribeVpc(ctx, ptr.Deref(vpc.VpcId, ""))
		if err != nil {
			return false, err
		}
		*vpc = *v
		anyInProgress := false
		for _, assoc := range vpc.CidrBlockAssociationSet {
			if assoc.CidrBlockState != nil {
				if assoc.CidrBlockState.State == ec2types.VpcCidrBlockStateCodeFailing ||
					assoc.CidrBlockState.State == ec2types.VpcCidrBlockStateCodeFailed {
					return false, fmt.Errorf("failed cidr block %q: %q", ptr.Deref(assoc.CidrBlock, ""), ptr.Deref(assoc.CidrBlockState.StatusMessage, ""))
				}
				if assoc.CidrBlockState.State == ec2types.VpcCidrBlockStateCodeAssociating ||
					assoc.CidrBlockState.State == ec2types.VpcCidrBlockStateCodeDisassociating {
					anyInProgress = true
				}
			}
		}
		return !anyInProgress, nil
	})
	if err != nil {
		return fmt.Errorf("error waiting vpc cidr assosiation to become available: %w", err)
	}
	return nil
}
