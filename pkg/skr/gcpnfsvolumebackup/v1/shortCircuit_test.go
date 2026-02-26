package v1

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type shortCircuitSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *shortCircuitSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *shortCircuitSuite) TestWhenBackupIsDeleting() {
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

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *shortCircuitSuite) TestWhenBackupIsReadyAndCapacityUpdate() {
	// isTimeForCapacityUpdate == true
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupReady
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().Add(-1 * time.Hour).Add(-1 * time.Minute)}
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	config.GcpConfig.GcpCapacityCheckInterval = time.Hour * 1
	err, _ctx := shortCircuitCompleted(ctx, state)

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
}

func (s *shortCircuitSuite) TestWhenBackupIsReadyAndNotCapacityUpdate() {
	// isTimeForCapacityUpdate == true
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupReady
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now()}
	obj.Status.FileStoreBackupLabels = map[string]string{
		client.ManagedByKey:             client.ManagedByValue,
		client.ScopeNameKey:             scope.Name,
		util.GcpLabelSkrVolumeName:      obj.Spec.Source.Volume.Name,
		util.GcpLabelSkrVolumeNamespace: obj.Spec.Source.Volume.Namespace,
		util.GcpLabelSkrBackupName:      obj.Name,
		util.GcpLabelSkrBackupNamespace: obj.Namespace,
		util.GcpLabelShootName:          scope.Spec.ShootName,
	}
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	state.Scope = &scope
	s.Nil(err)

	config.GcpConfig.GcpCapacityCheckInterval = time.Hour * 1
	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Equal(stopAndRequeueForCapacity(), err)
	s.Nil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupReady, fromK8s.Status.State)

}

func (s *shortCircuitSuite) TestWhenBackupIsInError() {
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

	err, _ctx := shortCircuitCompleted(ctx, state)

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

func (s *shortCircuitSuite) TestWhenBackupIsFailed() {
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

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
	s.Nil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err)

	s.Equal(v1beta1.GcpNfsBackupFailed, fromK8s.Status.State)
}

func (s *shortCircuitSuite) TestWhenBackupIsCreating() {
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

	err, _ctx := shortCircuitCompleted(ctx, state)

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

func TestShortCircuit(t *testing.T) {
	suite.Run(t, new(shortCircuitSuite))
}
