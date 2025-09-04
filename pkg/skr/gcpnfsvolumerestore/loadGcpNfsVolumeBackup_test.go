package gcpnfsvolumerestore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type loadGcpNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadGcpNfsVolumeBackupSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolumeRestore.DeepCopy()
	objDiffName.Spec.Source.Backup.Name = "diffName"

	factory, err := newTestStateFactoryWithObj(fakeHttpServer, objDiffName)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(objDiffName)
	s.Nil(err)
	err, _ctx := loadGcpNfsVolumeBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	s.Equal(ctx, _ctx)
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	// Remove the conditions from backup
	notReadyVolumeBackup := gcpNfsVolumeBackup.DeepCopy()
	notReadyVolumeBackup.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, notReadyVolumeBackup)
	s.Nil(err)
	err, _ = loadGcpNfsVolumeBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
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

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupReady() {
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
	err, ctx = loadGcpNfsVolumeBackup(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
}

func TestLoadGcpNfsVolumeBackupSuite(t *testing.T) {
	suite.Run(t, new(loadGcpNfsVolumeBackupSuite))
}
