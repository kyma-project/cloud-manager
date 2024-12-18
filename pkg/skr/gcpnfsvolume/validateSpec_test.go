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

func TestValidateSpec(t *testing.T) {
	suite.Run(t, new(validateSpecSuite))
}
