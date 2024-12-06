package gcpnfsvolume

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type validateSpecSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *validateSpecSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *validateSpecSuite) TestIpRangeWhenNotExist() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state := factory.newState()

	//Invoke validateIpRange
	err, _ = validateIpRange(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopAndForget, err)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(suite.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(suite.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionTypeError, nfsVol.Status.Conditions[0].Type)
	assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionReasonIpRangeNotFound, nfsVol.Status.Conditions[0].Reason)
}

func (suite *validateSpecSuite) TestIpRangeWhenNotReady() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add an IPRange to SKR.
	ipRange := cloudresourcesv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcpNfsVolume.Spec.IpRange.Name,
		},
		Spec: cloudresourcesv1beta1.IpRangeSpec{
			Cidr: kcpIpRange.Spec.Cidr,
		},
	}
	err = factory.skrCluster.K8sClient().Create(ctx, &ipRange)
	assert.Nil(suite.T(), err)

	state := factory.newState()

	//Invoke validateIpRange
	err, _ctx := validateIpRange(ctx, state)

	//validate expected return values
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(3*time.Second), err)
	assert.Nil(suite.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(suite.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(suite.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionTypeError, nfsVol.Status.Conditions[0].Type)
	assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionReasonIpRangeNotReady, nfsVol.Status.Conditions[0].Reason)
	assert.Equal(suite.T(), cloudresourcesv1beta1.GcpNfsVolumeError, nfsVol.Status.State)
}

func (suite *validateSpecSuite) TestIpRangeWhenReady() {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add an IPRange to SKR.
	ipRange := cloudresourcesv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcpNfsVolume.Spec.IpRange.Name,
		},
		Spec: cloudresourcesv1beta1.IpRangeSpec{
			Cidr: kcpIpRange.Spec.Cidr,
		},
		Status: cloudresourcesv1beta1.IpRangeStatus{
			Cidr: kcpIpRange.Spec.Cidr,
			Id:   "kcp-ip-range",
			Conditions: []metav1.Condition{
				{
					Type:    "Ready",
					Status:  "True",
					Reason:  "Ready",
					Message: "NFS is instance is ready",
				},
			},
		},
	}

	err = factory.skrCluster.K8sClient().Create(ctx, &ipRange)
	assert.Nil(suite.T(), err)

	state := factory.newState()

	//Invoke validateIpRange
	err, _ctx := validateIpRange(ctx, state)

	//validate expected return values
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(suite.T(), err)

	//Validate GcpNfsVolume status.
	assert.Nil(suite.T(), nfsVol.Status.Conditions)
}

func (suite *validateSpecSuite) checkTier(tier cloudresourcesv1beta1.GcpFileTier, valid bool) {
	factory, err := newTestStateFactory()
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Set the tier in GcpNfsVolume spec
	nfsVol := gcpNfsVolume.DeepCopy()
	nfsVol.Spec.Tier = tier

	err = factory.skrCluster.K8sClient().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	//Update Status and remove conditions.
	nfsVol.Status.Conditions = nil
	err = factory.skrCluster.K8sClient().Status().Update(ctx, nfsVol)
	assert.Nil(suite.T(), err)

	state := factory.newStateWith(nfsVol)

	//Invoke validateSpec
	err, _ = validateTier(ctx, state)

	//validate expected return values
	if valid {
		assert.Nil(suite.T(), err)
	} else {
		assert.Equal(suite.T(), composed.StopAndForget, err)
	}

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
		if tier == "UNSPECIFIED" {
			assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionReasonTierInvalid, nfsVol.Status.Conditions[0].Reason)

		} else {
			assert.Equal(suite.T(), cloudresourcesv1beta1.ConditionReasonTierLegacy, nfsVol.Status.Conditions[0].Reason)
		}
		assert.Equal(suite.T(), cloudresourcesv1beta1.GcpNfsVolumeError, nfsVol.Status.State)
	}

}

func (suite *validateSpecSuite) TestValidBasicHddTier() {
	suite.checkTier(cloudresourcesv1beta1.BASIC_HDD, true)
}

func (suite *validateSpecSuite) TestValidBasicSsdTier() {
	suite.checkTier(cloudresourcesv1beta1.BASIC_SSD, true)
}

func (suite *validateSpecSuite) TestValidZonalTier() {
	suite.checkTier(cloudresourcesv1beta1.ZONAL, true)
}

func (suite *validateSpecSuite) TestValidRegionalTier() {
	suite.checkTier(cloudresourcesv1beta1.REGIONAL, true)
}

func (suite *validateSpecSuite) TestInvalidLegacyStandardTier() {
	suite.checkTier(cloudresourcesv1beta1.STANDARD, false)
}

func (suite *validateSpecSuite) TestInvalidLegacyPremiumTier() {
	suite.checkTier(cloudresourcesv1beta1.PREMIUM, false)
}

func (suite *validateSpecSuite) TestInvalidLegacyHighScaleSsdTier() {
	suite.checkTier(cloudresourcesv1beta1.HIGH_SCALE_SSD, false)
}

func (suite *validateSpecSuite) TestInvalidLegacyEnterpriseTier() {
	suite.checkTier(cloudresourcesv1beta1.ENTERPRISE, false)
}

func (suite *validateSpecSuite) TestInvalidUnspecifiedTier() {
	suite.checkTier("UNSPECIFIED", false)
}

func TestValidateSpec(t *testing.T) {
	suite.Run(t, new(validateSpecSuite))
}
