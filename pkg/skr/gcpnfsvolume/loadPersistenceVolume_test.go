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

type loadPersistenceVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadPersistenceVolumeSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *loadPersistenceVolumeSuite) TestWithMatchingPV() {
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

	//Add an PV to SKR.
	pv := pvGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(s.T(), err)

	err, _ctx := loadPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the IpRange object
	assert.NotNil(s.T(), state.PV)
}

func (s *loadPersistenceVolumeSuite) TestWithNotMatchingPV() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add an PV to SKR.
	pv := pvGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
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

	err, _ctx := loadPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the IpRange object
	assert.Nil(s.T(), state.PV)
}

func (s *loadPersistenceVolumeSuite) TestWithMultipleMatchingIpRanges() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Add an PV to SKR.
	pv := pvGcpNfsVolume.DeepCopy()
	err = factory.skrCluster.K8sClient().Create(ctx, pv)
	assert.Nil(s.T(), err)

	//Add another PV to SKR.
	pv2 := pvGcpNfsVolume.DeepCopy()
	pv2.Name = "test-pv-2"
	err = factory.skrCluster.K8sClient().Create(ctx, pv2)
	assert.Nil(s.T(), err)

	//Get state object with GcpNfsVolume
	state, err := factory.newState()
	assert.Nil(s.T(), err)

	err, _ctx := loadPersistenceVolume(ctx, state)

	//validate expected return values
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), _ctx)

	//Validate the IpRange object
	assert.NotNil(s.T(), state.PV)
}

func TestLoadPersistenceVolumeSuite(t *testing.T) {
	suite.Run(t, new(loadPersistenceVolumeSuite))
}
