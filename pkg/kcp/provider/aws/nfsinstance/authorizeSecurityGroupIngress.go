package nfsinstance

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func authorizeSecurityGroupIngress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	toPort := int32(2049)

	hasPods := state.Scope().Spec.Scope.Aws.Network.Pods == ""
	hasNet := state.Scope().Spec.Scope.Aws.Network.VPC.CIDR == ""

	for _, perm := range state.securityGroup.IpPermissions {
		if pointer.Int32Deref(perm.ToPort, 0) != toPort {
			continue
		}
		if pointer.StringDeref(perm.IpProtocol, "") != "tcp" {
			continue
		}
		for _, rng := range perm.IpRanges {
			if pointer.StringDeref(rng.CidrIp, "") == state.Scope().Spec.Scope.Aws.Network.VPC.CIDR {
				hasNet = true
			}
			if pointer.StringDeref(rng.CidrIp, "") == state.Scope().Spec.Scope.Aws.Network.Pods {
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
			IpProtocol: pointer.String("tcp"),
			FromPort:   pointer.Int32(toPort),
			ToPort:     pointer.Int32(toPort),
			IpRanges: []ec2Types.IpRange{
				{
					CidrIp: pointer.String(state.Scope().Spec.Scope.Aws.Network.Pods),
				},
			},
		})
	}
	if !hasNet {
		logger.Info("Adding vpc cidr to the NFS security group")
		permissions = append(permissions, ec2Types.IpPermission{
			IpProtocol: pointer.String("tcp"),
			FromPort:   pointer.Int32(toPort),
			ToPort:     pointer.Int32(toPort),
			IpRanges: []ec2Types.IpRange{
				{
					CidrIp: pointer.String(state.Scope().Spec.Scope.Aws.Network.VPC.CIDR),
				},
			},
		})
	}

	if len(permissions) == 0 {
		return nil, nil
	}

	err := state.awsClient.AuthorizeSecurityGroupIngress(ctx, state.securityGroupId, permissions)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error adding security group ingress", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
