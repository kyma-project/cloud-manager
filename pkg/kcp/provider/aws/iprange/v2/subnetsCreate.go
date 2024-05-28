package v2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awserrorhandling "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/errorhandling"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/pointer"
)

func subnetsCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	count := len(state.ObjAsIpRange().Status.Ranges)

	rangeMap := make(map[string]interface{}, count)
	for _, r := range state.ObjAsIpRange().Status.Ranges {
		rangeMap[r] = nil
	}

	zoneMap := make(map[string]interface{}, count)
	for _, z := range state.Scope().Spec.Scope.Aws.Network.Zones {
		zoneMap[z.Name] = nil
	}

	foundCount := 0

	for _, subnet := range state.cloudResourceSubnets {
		subnetName := awsutil.GetEc2TagValue(subnet.Tags, "Name")
		zoneValue := pointer.StringDeref(subnet.AvailabilityZone, "")
		rangeValue := pointer.StringDeref(subnet.CidrBlock, "")
		key := fmt.Sprintf("%s/%s", zoneValue, rangeValue)
		if len(key) <= 3 {
			logger.
				WithValues(
					"awsAccount", state.Scope().Spec.Scope.Aws.AccountId,
					"subnetId", subnet.SubnetId,
					"subnetName", subnetName,
					"zone", zoneValue,
					"cidr", rangeValue,
				).
				Info("Subnet missing availability zone and/or cidr block!")
			continue
		}

		logger.
			WithValues(
				"zone", zoneValue,
				"range", rangeValue,
				"subnetId", subnet.SubnetId,
				"subnetName", subnetName,
			).
			Info("Zone already exist")

		delete(zoneMap, zoneValue)
		delete(rangeMap, rangeValue)
		foundCount++
	}

	indexMap := make(map[string]int, count)
	for i, z := range state.Scope().Spec.Scope.Aws.Network.Zones {
		indexMap[z.Name] = i
	}

	anyCreated := false
	zones := pie.Keys(zoneMap)
	for i, rng := range pie.Keys(rangeMap) {
		zn := zones[i]
		logger := logger.
			WithValues(
				"zone", zn,
				"range", rng,
			)
		logger.Info("Creating subnet")

		idx := indexMap[zn]
		subnet, err := state.awsClient.CreateSubnet(ctx, aws.ToString(state.vpc.VpcId), zn, rng, awsutil.Ec2Tags(
			"Name", fmt.Sprintf("%s-%d", state.ObjAsIpRange().Name, idx),
			common.TagCloudManagerName, state.Name().String(),
			common.TagCloudManagerRemoteName, state.ObjAsIpRange().Spec.RemoteRef.String(),
			common.TagScope, state.ObjAsIpRange().Spec.Scope.Name,
			tagKey, "1",
		))
		if x := awserrorhandling.HandleError(ctx, err, state, "KCP IpRange on create subnet",
			cloudcontrolv1beta1.ReasonUnknown, "Failed creating subnet"); x != nil {
			return x, nil
		}
		anyCreated = true

		logger.WithValues("subnetId", subnet.SubnetId).Info("Subnet created")

		state.ObjAsIpRange().Status.Subnets = append(state.ObjAsIpRange().Status.Subnets, cloudcontrolv1beta1.IpRangeSubnet{
			Id:    pointer.StringDeref(subnet.SubnetId, ""),
			Zone:  pointer.StringDeref(subnet.AvailabilityZone, ""),
			Range: pointer.StringDeref(subnet.CidrBlock, ""),
		})

		x, _ := composed.PatchStatus(state.ObjAsIpRange()).
			SuccessErrorNil().
			Run(ctx, state)
		if x != nil {
			return x, ctx
		}
	}

	if anyCreated {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}

	return nil, nil
}
