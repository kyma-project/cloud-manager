package redisinstance

import (
	"context"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func authorizeSecurityGroupIngress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	toPort := int32(6379)

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
			return nil, ctx
		}
	}

	var permissions []ec2types.IpPermission

	if !hasPods {
		logger.Info("Adding pod cidr to the Redis security group")
		permissions = append(permissions, ec2types.IpPermission{
			IpProtocol: new("tcp"),
			FromPort:   new(toPort),
			ToPort:     new(toPort),
			IpRanges: []ec2types.IpRange{
				{
					CidrIp: new(state.Scope().Spec.Scope.Aws.Network.Pods),
				},
			},
		})
	}
	if !hasNet {
		logger.Info("Adding vpc cidr to the Redis security group")
		permissions = append(permissions, ec2types.IpPermission{
			IpProtocol: new("tcp"),
			FromPort:   new(toPort),
			ToPort:     new(toPort),
			IpRanges: []ec2types.IpRange{
				{
					CidrIp: new(state.Scope().Spec.Scope.Aws.Network.VPC.CIDR),
				},
			},
		})
	}

	if len(permissions) == 0 {
		return nil, ctx
	}

	err := state.awsClient.AuthorizeElastiCacheSecurityGroupIngress(ctx, state.securityGroupId, permissions)
	if err != nil {
		if awsmeta.IsErrorRetryable(err) {
			return awsmeta.LogErrorAndReturn(err, "Error adding security group ingress", ctx)
		}
		logger.Error(err, "Error adding security group ingress")
		redisInstance := state.ObjAsRedisInstance()
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to authorize security group ingress",
		})
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		updateErr := state.UpdateObjStatus(ctx)
		if updateErr != nil {
			return composed.LogErrorAndReturn(updateErr,
				"Error updating RedisInstance status due failed security group ingress authorization",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
