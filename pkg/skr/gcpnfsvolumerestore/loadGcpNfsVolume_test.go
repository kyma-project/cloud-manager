package gcpnfsvolumerestore

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type loadGcpNfsVolumeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadGcpNfsVolumeSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *loadGcpNfsVolumeSuite) TestVolumeNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolumeRestore.DeepCopy()
	objDiffName.Spec.Destination.Volume.Name = "diffName"

	factory, err := newTestStateFactoryWithObj(fakeHttpServer, objDiffName)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(objDiffName)
	s.Nil(err)
	err, _ctx := loadGcpNfsVolume(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
	s.Equal(ctx, _ctx)
}

func (s *loadGcpNfsVolumeSuite) TestVolumeNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	// Remove the conditions from volume
	notReadyVolume := gcpNfsVolume.DeepCopy()
	notReadyVolume.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, notReadyVolume)
	s.Nil(err)
	err, _ = loadGcpNfsVolume(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeRestore.Name,
			Namespace: gcpNfsVolumeRestore.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.JobStateError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (s *loadGcpNfsVolumeSuite) TestVolumeReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	err, ctx = loadGcpNfsVolume(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
}

func TestLoadGcpNfsVolumeSuite(t *testing.T) {
	suite.Run(t, new(loadGcpNfsVolumeSuite))
}
