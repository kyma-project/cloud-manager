package gcpnfsvolumebackup

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type shortCircuitSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *shortCircuitSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *shortCircuitSuite) TestWhenBackupIsDeleting() {
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

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *shortCircuitSuite) TestWhenBackupIsReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupReady
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)
	suite.Nil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.GcpNfsBackupReady, fromK8s.Status.State)
}

func (suite *shortCircuitSuite) TestWhenBackupIsInError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupError
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.GcpNfsBackupError, fromK8s.Status.State)
}

func (suite *shortCircuitSuite) TestWhenBackupIsFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupFailed
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopAndForget, err)
	suite.Nil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.GcpNfsBackupFailed, fromK8s.Status.State)
}

func (suite *shortCircuitSuite) TestWhenBackupIsCreating() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupCreating
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := shortCircuitCompleted(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Equal(ctx, _ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err)

	suite.Equal(v1beta1.GcpNfsBackupCreating, fromK8s.Status.State)
}

func TestShortCircuit(t *testing.T) {
	suite.Run(t, new(shortCircuitSuite))
}
