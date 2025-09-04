package nfsinstance

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type validationsSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *validationsSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *validationsSuite) TestIsValidCapacity() {
	testMinMax(s, v1beta1.BASIC_HDD, 1023, 66000, 1024, 65000)
	testMinMax(s, v1beta1.BASIC_SSD, 2559, 66000, 2560, 65000)
	testMinMax(s, v1beta1.ZONAL, 1023, 10241, 1024, 10240)
	testMinMax(s, v1beta1.REGIONAL, 1023, 10241, 1024, 10240)
}

func testMinMax(s *validationsSuite, tier v1beta1.GcpFileTier, minErr, maxErr, minOk, maxOk int) {
	valid, err := IsValidCapacity(tier, minErr)
	assert.False(s.T(), valid)
	assert.NotNil(s.T(), err)
	valid, err = IsValidCapacity(tier, maxErr)
	assert.False(s.T(), valid)
	assert.NotNil(s.T(), err)
	valid, err = IsValidCapacity(tier, minOk)
	assert.True(s.T(), valid)
	assert.Nil(s.T(), err)
	valid, err = IsValidCapacity(tier, maxOk)
	assert.True(s.T(), valid)
	assert.Nil(s.T(), err)
}

func (s *validationsSuite) TestCanScaleDown() {
	assert.False(s.T(), CanScaleDown(v1beta1.BASIC_HDD))
	assert.False(s.T(), CanScaleDown(v1beta1.BASIC_SSD))
	assert.True(s.T(), CanScaleDown(v1beta1.ZONAL))
	assert.True(s.T(), CanScaleDown(v1beta1.REGIONAL))
}
func TestValidations(t *testing.T) {
	suite.Run(t, new(validationsSuite))
}
