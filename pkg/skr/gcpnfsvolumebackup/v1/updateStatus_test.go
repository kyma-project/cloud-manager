package v1

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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

func (s *updateStatusSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *updateStatusSuite) TestDeletingBackupExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	state.fileBackup = &file.Backup{}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	s.Equal(err, composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime))
	s.Nil(_ctx)
}

func (s *updateStatusSuite) TestDeletingBackupNotExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	s.Equal(err, composed.StopAndForget)
	s.Nil(_ctx)
}

func (s *updateStatusSuite) TestReadyAndBackupExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	state.fileBackup = &file.Backup{State: "READY"}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupReady, fromK8s.Status.State)
	s.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	s.Equal(cloudcontrolv1beta1.ConditionTypeReady, fromK8s.Status.Conditions[0].Type)
}

func (s *updateStatusSuite) TestNotReadyAndBackupReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))

	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Update the status with empty conditions
	obj.Status.Conditions = []metav1.Condition{}
	err = state.SkrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	state.fileBackup = &file.Backup{State: "READY"}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupReady, fromK8s.Status.State)
	s.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	s.Equal(cloudcontrolv1beta1.ConditionTypeReady, fromK8s.Status.Conditions[0].Type)
}

func (s *updateStatusSuite) TestNotReadyAndBackupNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))

	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Update the status with empty conditions
	obj.Status.Conditions = []metav1.Condition{}
	err = state.SkrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	state.fileBackup = &file.Backup{State: "CREATING"}
	err, _ctx := updateStatus(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
	s.Nil(_ctx)
}

func TestUpdateStatus(t *testing.T) {
	suite.Run(t, new(updateStatusSuite))
}
