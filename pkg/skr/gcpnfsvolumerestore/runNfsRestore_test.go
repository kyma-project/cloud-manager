package gcpnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	"io"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type runNfsRestoreSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *runNfsRestoreSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *runNfsRestoreSuite) TestRunNfsRestoreAlreadyOpId() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj.Status.OpIdentifier = "op-123"
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	err, ctx = runNfsRestore(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func (suite *runNfsRestoreSuite) TestRunNfsRestoreJobCompleted() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj.Status.State = cloudresourcesv1beta1.JobStateFailed
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	err, ctx = runNfsRestore(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)

	// Done
	obj.Status.State = cloudresourcesv1beta1.JobStateDone
	state, err = factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	err, ctx = runNfsRestore(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func (suite *runNfsRestoreSuite) TestRunNfsRestoreDeletionTimestamp() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	deletingObj := deletingGpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, deletingObj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(deletingObj)
	assert.Nil(suite.T(), err)
	err, ctx = runNfsRestore(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func (suite *runNfsRestoreSuite) TestRunNfsRestoreSubmitFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Restore submission failed.", http.StatusInternalServerError)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolumeBackup = gcpNfsVolumeBackup.DeepCopy()
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	assert.Nil(suite.T(), err)
	err, _ctx := runNfsRestore(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
	assert.Equal(suite.T(), ctx, _ctx)
}

func (suite *runNfsRestoreSuite) TestRunNfsRestoreSubmitted() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opResp := file.Operation{
			Name: "op-123",
			Done: false,
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(suite.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
		body, err := io.ReadAll(r.Body)
		assert.Nil(suite.T(), err)
		//create filestore instance from byte[] and check if it is equal to the expected filestore instance
		obj := &file.RestoreInstanceRequest{}
		err = json.Unmarshal(body, obj)
		assert.Nil(suite.T(), err)
		suite.Contains(obj.SourceBackup, "projects/test-project/locations/us-west1/backups/cm-"+gcpNfsVolumeBackup.Status.Id)
		suite.Equal("/v1/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance:restore", r.URL.Path)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolumeBackup = gcpNfsVolumeBackup.DeepCopy()
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	assert.Nil(suite.T(), err)
	err, _ctx := runNfsRestore(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpOperationWaitTime), err)
	assert.Equal(suite.T(), ctx, _ctx)
	fromK8s := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "op-123", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), cloudresourcesv1beta1.JobStateInProgress, fromK8s.Status.State)
}

func TestRunNfsRestore(t *testing.T) {
	suite.Run(t, new(runNfsRestoreSuite))
}
