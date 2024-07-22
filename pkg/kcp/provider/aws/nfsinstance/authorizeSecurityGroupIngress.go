package nfsinstance

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func authorizeSecurityGroupIngress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	toPort := int32(2049)

	hasPods := state.Scope().Spec.Scope.Aws.Network.Pods == ""
	hasNet := state.Scope().Spec.Scope.Aws.Network.VPC.CIDR == ""

	for _, perm := range state.securityGroup.IpPermissions {
		if ptr.Deref(perm.ToPort, 0) != toPort {
			continue
		}
		if ptr.Deref(perm.IpProtocol, "") != "tcp" {
			continue
		}
		for _, rng := range perm.IpRanges {
			if ptr.Deref(rng.CidrIp, "") == state.Scope().Spec.Scope.Aws.Network.VPC.CIDR {
				hasNet = true
			}
			if ptr.Deref(rng.CidrIp, "") == state.Scope().Spec.Scope.Aws.Network.Pods {
				hasPods = true
			}
		}
		if hasPods && hasNet {
			return nil, nil
		}
	}

	var permissions []ec2Types.IpPermission

	if !hasPods {
		logger.Info("Adding pod cidr to the NFS security group")
		permissions = append(permissions, ec2Types.IpPermission{
			IpProtocol: ptr.To("tcp"),
			FromPort:   ptr.To(toPort),
			ToPort:     ptr.To(toPort),
			IpRanges: []ec2Types.IpRange{
				{
					CidrIp: ptr.To(state.Scope().Spec.Scope.Aws.Network.Pods),
				},
			},
		})
	}
	if !hasNet {
		logger.Info("Adding vpc cidr to the NFS security group")
		permissions = append(permissions, ec2Types.IpPermission{
			IpProtocol: ptr.To("tcp"),
			FromPort:   ptr.To(toPort),
			ToPort:     ptr.To(toPort),
			IpRanges: []ec2Types.IpRange{
				{
					CidrIp: ptr.To(state.Scope().Spec.Scope.Aws.Network.VPC.CIDR),
				},
			},
		})
	}

	if len(permissions) == 0 {
		return nil, nil
	}

	err := state.awsClient.AuthorizeSecurityGroupIngress(ctx, state.securityGroupId, permissions)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error adding security group ingress", ctx)
	}

	return composed.StopWithRequeue, nil
}
