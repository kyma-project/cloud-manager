package iprange

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
	"k8s.io/utils/pointer"
)

func createSubnets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	count := len(state.IpRange().Status.Ranges)

	rangeMap := make(map[string]interface{}, count)
	for _, r := range state.IpRange().Status.Ranges {
		rangeMap[r] = nil
	}

	zoneMap := make(map[string]interface{}, count)
	for _, z := range state.Scope().Spec.Scope.Aws.Network.Zones {
		zoneMap[z.Name] = nil
	}

	foundCount := 0

	for _, subnet := range state.cloudResourceSubnets {
		subnetName := getTagValue(subnet.Tags, "Name")
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
		subnet, err := state.networkClient.CreateSubnet(ctx, aws.ToString(state.vpc.VpcId), zn, rng, []ec2Types.Tag{
			{
				Key:   pointer.String("Name"),
				Value: pointer.String(fmt.Sprintf("cloudresources-%d", idx)),
			},
			{
				Key:   pointer.String(tagKey),
				Value: pointer.String("1"),
			},
		})
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error creating subnet", composed.StopWithRequeue, nil)
		}
		logger.WithValues("subnetId", subnet.SubnetId).Info("Subnet created")
	}

	return nil, nil
}
