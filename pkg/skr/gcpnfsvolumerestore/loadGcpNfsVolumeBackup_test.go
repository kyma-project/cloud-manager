package gcpnfsvolumerestore

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"testing"
)

type loadGcpNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadGcpNfsVolumeBackupSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolumeRestore.DeepCopy()
	objDiffName.Spec.Source.Backup.Name = "diffName"

	factory, err := newTestStateFactoryWithObj(fakeHttpServer, objDiffName)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(objDiffName)
	suite.Nil(err)
	err, _ctx := loadGcpNfsVolumeBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	suite.Equal(ctx, _ctx)
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	// Remove the conditions from backup
	notReadyVolumeBackup := gcpNfsVolumeBackup.DeepCopy()
	notReadyVolumeBackup.Status.Conditions = []metav1.Condition{}
	err = factory.skrCluster.K8sClient().Status().Update(ctx, notReadyVolumeBackup)
	suite.Nil(err)
	err, _ = loadGcpNfsVolumeBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeRestore.Name,
			Namespace: gcpNfsVolumeRestore.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.JobStateError, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	err, ctx = loadGcpNfsVolumeBackup(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func TestLoadGcpNfsVolumeBackupSuite(t *testing.T) {
	suite.Run(t, new(loadGcpNfsVolumeBackupSuite))
}
