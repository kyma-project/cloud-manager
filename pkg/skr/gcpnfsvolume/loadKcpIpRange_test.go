package gcpnfsvolume

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type loadKcpIpRangeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadKcpIpRangeSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadKcpIpRangeSuite) TestWithMatchingIpRange() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newState()
	assert.Nil(s.T(), err)
	state.SkrIpRange = skrIpRange.DeepCopy()

	//Add an IPRange to KCP.
	ipRange := kcpIpRange.DeepCopy()
	err = factory.kcpCluster.K8sClient().Create(ctx, ipRange)
	assert.Nil(s.T(), err)

	err, _ctx := loadKcpIpRange(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the IpRange object
	assert.NotNil(s.T(), state.KcpIpRange)
}

func (s *loadKcpIpRangeSuite) TestWithNotMatchingIpRange() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add an IPRange to KCP.
	ipRange := kcpIpRange.DeepCopy()
	err = factory.kcpCluster.K8sClient().Create(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Get state object with GcpNfsVolume
	nfsVol := cloudresourcesv1beta1.GcpNfsVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "not-matching-gcp-nfs-volume",
			Namespace: "test",
		},
	}
	state, err := factory.newStateWith(&nfsVol)
	s.Nil(err)
	state.SetSkrIpRange(&cloudresourcesv1beta1.IpRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "not-matching-ip-range",
			Namespace: "test",
		},
	})

	err, _ctx := loadKcpIpRange(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the IpRange object
	assert.Nil(s.T(), state.KcpIpRange)
}

func (s *loadKcpIpRangeSuite) TestWithMultipleMatchingIpRanges() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add an IPRange to KCP.
	ipRange := kcpIpRange.DeepCopy()
	err = factory.kcpCluster.K8sClient().Create(ctx, ipRange)
	assert.Nil(s.T(), err)

	//Add another IPRange to KCP.
	ipRange2 := kcpIpRange.DeepCopy()
	ipRange2.Name = "test-ip-range-2"
	err = factory.kcpCluster.K8sClient().Create(ctx, ipRange2)
	assert.Nil(s.T(), err)

	//Get state object with GcpNfsVolume
	state, err := factory.newState()
	assert.Nil(s.T(), err)
	state.SkrIpRange = skrIpRange.DeepCopy()

	err, _ctx := loadKcpIpRange(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the IpRange object
	assert.NotNil(s.T(), state.KcpIpRange)
}

func TestLoadKcpIpRangeSuite(t *testing.T) {
	suite.Run(t, new(loadKcpIpRangeSuite))
}
