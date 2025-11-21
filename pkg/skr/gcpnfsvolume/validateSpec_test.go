package gcpnfsvolume

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func (s *validateSpecSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *validateSpecSuite) TestIpRangeWhenNotExist() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newState()
	assert.Nil(s.T(), err)

	//Invoke validateIpRange
	err, _ = validateIpRange(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), composed.StopAndForget, err)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(s.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudresourcesv1beta1.ConditionTypeError, nfsVol.Status.Conditions[0].Type)
	assert.Equal(s.T(), cloudresourcesv1beta1.ConditionReasonIpRangeNotFound, nfsVol.Status.Conditions[0].Reason)
}

func (s *validateSpecSuite) TestIpRangeWhenNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	state, err := factory.newState()
	assert.Nil(s.T(), err)

	//Invoke validateIpRange
	err, _ctx := validateIpRange(ctx, state)

	//validate expected return values
	assert.Equal(s.T(), composed.StopWithRequeueDelay(3*time.Second), err)
	assert.Nil(s.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Equal(s.T(), 1, len(nfsVol.Status.Conditions))
	assert.Equal(s.T(), metav1.ConditionTrue, nfsVol.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudresourcesv1beta1.ConditionTypeError, nfsVol.Status.Conditions[0].Type)
	assert.Equal(s.T(), cloudresourcesv1beta1.ConditionReasonIpRangeNotReady, nfsVol.Status.Conditions[0].Reason)
	assert.Equal(s.T(), cloudresourcesv1beta1.GcpNfsVolumeError, nfsVol.Status.State)
}

func (s *validateSpecSuite) TestIpRangeWhenReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)

	state, err := factory.newState()
	assert.Nil(s.T(), err)

	//Invoke validateIpRange
	err, _ctx := validateIpRange(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Get the modified GcpNfsVolume object
	nfsVol := &cloudresourcesv1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name, Namespace: gcpNfsVolume.Namespace}, nfsVol)
	assert.Nil(s.T(), err)

	//Validate GcpNfsVolume status.
	assert.Nil(s.T(), nfsVol.Status.Conditions)
}

func TestValidateSpec(t *testing.T) {
	suite.Run(t, new(validateSpecSuite))
}
