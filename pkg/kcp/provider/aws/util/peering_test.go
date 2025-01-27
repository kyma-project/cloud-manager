package util

import (
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	shootName string = "c-123123"
)

func TestAwsRouteTableUpdateStrategyAuto(t *testing.T) {
	assert.True(t, ShouldUpdateRouteTable(Ec2Tags(), v1beta1.AwsRouteTableUpdateStrategyAuto, shootName))
}

func TestAwsRouteTableUpdateStrategyMatched(t *testing.T) {
	assert.True(t, ShouldUpdateRouteTable(Ec2Tags(shootName), v1beta1.AwsRouteTableUpdateStrategyMatched, shootName))
}

func TestAwsRouteTableUpdateStrategyMatchedNoTag(t *testing.T) {
	assert.False(t, ShouldUpdateRouteTable(Ec2Tags(), v1beta1.AwsRouteTableUpdateStrategyMatched, shootName))
}

func TestAwsRouteTableUpdateStrategyUnmatched(t *testing.T) {
	assert.True(t, ShouldUpdateRouteTable(Ec2Tags(), v1beta1.AwsRouteTableUpdateStrategyUnmatched, shootName))
}

func TestAwsRouteTableUpdateStrategyUnmatchedHasTag(t *testing.T) {
	assert.False(t, ShouldUpdateRouteTable(Ec2Tags(shootName), v1beta1.AwsRouteTableUpdateStrategyUnmatched, shootName))
}

func TestShouldDeleteRoutesWhenDeleteRemotePeeringTrueAndUpdateStrategyAuto(t *testing.T) {
	assert.True(t, ShouldDeleteRoutes(true, v1beta1.AwsRouteTableUpdateStrategyAuto))
}
func TestShouldDeleteRoutesWhenDeleteRemotePeeringTrueAndUpdateStrategyNone(t *testing.T) {
	assert.True(t, ShouldDeleteRoutes(true, v1beta1.AwsRouteTableUpdateStrategyNone))
}

func TestShouldDeleteRoutesWhenDeleteRemotePeeringFalseAndUpdateStrategyAuto(t *testing.T) {
	assert.True(t, ShouldDeleteRoutes(false, v1beta1.AwsRouteTableUpdateStrategyAuto))
}
func TestShouldDeleteRoutesWhenDeleteRemotePeeringFalseAndUpdateStrategyNone(t *testing.T) {
	assert.False(t, ShouldDeleteRoutes(false, v1beta1.AwsRouteTableUpdateStrategyNone))
}
