package gcpnfsvolumebackup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type markFailedSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *markFailedSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *markFailedSuite) TestWhenBackupIsDeleting() {
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

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *markFailedSuite) TestWhenBackupIsReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupReady
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Equal(ctx, _ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupReady, fromK8s.Status.State)
}

func (s *markFailedSuite) TestWhenBackupIsFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupFailed
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Equal(ctx, _ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupFailed, fromK8s.Status.State)
}

func (s *markFailedSuite) TestWhenBackupIsCreating() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupCreating
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Equal(ctx, _ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupCreating, fromK8s.Status.State)
}

func (s *markFailedSuite) TestWhenBackupIsLatestAndInError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupError
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Equal(ctx, _ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupError, fromK8s.Status.State)
}

func (s *markFailedSuite) TestWhenBackupIsNotLatestAndInError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	labels := map[string]string{
		v1beta1.LabelScheduleName:      "test-schedule",
		v1beta1.LabelScheduleNamespace: "test",
	}

	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.CreationTimestamp = metav1.Time{Time: time.Now().Add(-1 * time.Minute)}
	obj.Labels = labels
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj.Status.State = v1beta1.GcpNfsBackupError
	err = factory.skrCluster.K8sClient().Status().Update(ctx, obj)
	s.Nil(err)

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Create another backup object for the same schedule
	obj2 := gcpNfsVolumeBackup.DeepCopy()
	obj2.Name = "test-backup-02"
	obj2.Namespace = "test"
	obj2.CreationTimestamp = metav1.Time{Time: time.Now()}
	obj2.Labels = labels
	obj2.Status.State = v1beta1.GcpNfsBackupReady
	err = factory.skrCluster.K8sClient().Create(ctx, obj2)
	s.Nil(err)

	err, _ctx := markFailed(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
	s.Equal(ctx, _ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupFailed, fromK8s.Status.State)
	s.Equal(v1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
	s.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	s.Equal(v1beta1.ReasonBackupFailed, fromK8s.Status.Conditions[0].Reason)
}

func TestMarkFailed(t *testing.T) {
	suite.Run(t, new(markFailedSuite))
}
