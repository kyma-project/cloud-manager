package mock

import (
	"context"

	alicloudiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/iprange/client"
	alicloudvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcnetwork/client"
)

// iprangeClientView adapts vpcStore to alicloudiprangeclient.Client.
type iprangeClientView struct{ *vpcStore }

var _ alicloudiprangeclient.Client = (*iprangeClientView)(nil)

func (c *iprangeClientView) DescribeVpcs(ctx context.Context, name string) ([]alicloudiprangeclient.VpcInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	raw := c.describeVpcsRaw(name)
	out := make([]alicloudiprangeclient.VpcInfo, 0, len(raw))
	for _, v := range raw {
		out = append(out, alicloudiprangeclient.VpcInfo{VpcId: v.VpcId, VpcName: v.VpcName, CidrBlock: v.CidrBlock, Status: v.Status})
	}
	return out, nil
}

// vpcnetworkClientView adapts vpcStore to alicloudvpcnetworkclient.Client.
type vpcnetworkClientView struct{ *vpcStore }

var _ alicloudvpcnetworkclient.Client = (*vpcnetworkClientView)(nil)

func (c *vpcnetworkClientView) DescribeVpcs(ctx context.Context, name string) ([]alicloudvpcnetworkclient.VpcInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	raw := c.describeVpcsRaw(name)
	out := make([]alicloudvpcnetworkclient.VpcInfo, 0, len(raw))
	for _, v := range raw {
		out = append(out, alicloudvpcnetworkclient.VpcInfo{VpcId: v.VpcId, VpcName: v.VpcName, CidrBlock: v.CidrBlock, Status: v.Status})
	}
	return out, nil
}
