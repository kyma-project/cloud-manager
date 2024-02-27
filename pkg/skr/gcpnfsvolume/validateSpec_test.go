package gcpnfsvolume

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type validateSpecSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *validateSpecSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *validateSpecSuite) TestDoNotValidateCapacity() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state := factory.newState()

	//Invoke validateSpec
	err, _ctx := validateCapacity(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)
}

func (suite *validateSpecSuite) TestValidCapacityBasicHDD() {
	suite.checkCapacity(2048, cloudresourcesv1beta1.BASIC_HDD, true)
}

func (suite *validateSpecSuite) TestLowCapacityBasicHDD() {
	suite.checkCapacity(600, cloudresourcesv1beta1.BASIC_HDD, false)
}

func (suite *validateSpecSuite) TestHighCapacityBasicHDD() {
	suite.checkCapacity(66000, cloudresourcesv1beta1.BASIC_HDD, false)
}

func (suite *validateSpecSuite) checkCapacity(capacityGb int, tier cloudresourcesv1beta1.GcpFileTier, valid bool) {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Set the capacity in GcpNfsVolume spec
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Spec.Tier = tier
	nfsVol.Spec.CapacityGb = capacityGb

	err = factory.skrCluster.K8sClient().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	//Update Status and remove conditions.
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	state := factory.newStateWith(nfsVol)

	//Invoke validateSpec
	err, _ctx := validateCapacity(ctx, state)

	//validate expected return values
	if valid {
		assert.Nil(suite.T(), err)
	} else {
		assert.Equal(suite.T(), err, composed.StopAndForget)
	}
	assert.Nil(suite.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol = &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	if valid {
		assert.Equal(suite.T(), 0, len(nfsVol.Status.Conditions))
	} else {
		assert.Equal(suite.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
		assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionTypeError, nfsVol.Status.Conditions[0].Type)
		assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionReasonCapacityInvalid, nfsVol.Status.Conditions[0].Reason)
	}

}

func TestValidateSpec(t *testing.T) {
	suite.Run(t, new(validateSpecSuite))
}
