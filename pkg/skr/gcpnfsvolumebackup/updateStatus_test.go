package gcpnfsvolumebackup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type updateStatusSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *updateStatusSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *updateStatusSuite) TestDeletingBackupExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	state.fileBackup = &file.Backup{}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	suite.Equal(err, composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime))
	suite.Nil(_ctx)
}

func (suite *updateStatusSuite) TestDeletingBackupNotExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	suite.Equal(err, composed.StopAndForget)
	suite.Nil(_ctx)
}

func (suite *updateStatusSuite) TestReadyAndBackupExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	state.fileBackup = &file.Backup{State: "READY"}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.GcpNfsBackupReady, fromK8s.Status.State)
	suite.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	suite.Equal(cloudcontrolv1beta1.ConditionTypeReady, fromK8s.Status.Conditions[0].Type)
}

func (suite *updateStatusSuite) TestNotReadyAndBackupReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))

	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the status with empty conditions
	obj.Status.Conditions = []metav1.Condition{}
	err = state.SkrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	state.fileBackup = &file.Backup{State: "READY"}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.GcpNfsBackupReady, fromK8s.Status.State)
	suite.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	suite.Equal(cloudcontrolv1beta1.ConditionTypeReady, fromK8s.Status.Conditions[0].Type)
}

func (suite *updateStatusSuite) TestNotReadyAndBackupNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))

	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Update the status with empty conditions
	obj.Status.Conditions = []metav1.Condition{}
	err = state.SkrCluster.K8sClient().Status().Update(ctx, obj)
	suite.Nil(err)

	state.fileBackup = &file.Backup{State: "CREATING"}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	suite.Nil(_ctx)
}

func TestUpdateStatus(t *testing.T) {
	suite.Run(t, new(updateStatusSuite))
}
