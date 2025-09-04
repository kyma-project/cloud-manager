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
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type checkBackupOperationSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *checkBackupOperationSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *checkBackupOperationSuite) TestCheckBackupOperationNoOpId() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	err, ctx = checkBackupOperation(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
}

func (s *checkBackupOperationSuite) TestCheckBackupOperationErrorNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.OpIdentifier = "op-123"
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.Scope = &scope
	err, ctx = checkBackupOperation(ctx, state)
	assert.Equal(s.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), v1beta1.GcpNfsBackupError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (s *checkBackupOperationSuite) TestCheckBackupOperationOtherError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer fakeHttpServer.Close()

	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj.Status.OpIdentifier = "op-123"
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.Scope = &scope
	err, ctx = checkBackupOperation(ctx, state)
	assert.Equal(s.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "op-123", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), v1beta1.GcpNfsBackupError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (s *checkBackupOperationSuite) TestCheckBackupOperationNotCompleted() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opResp := file.Operation{
			Name: "op-123",
			Done: false,
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.OpIdentifier = "op-123"
	obj.Status.State = v1beta1.GcpNfsBackupCreating
	obj.Status.Conditions = nil
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.Scope = &scope
	err, ctx = checkBackupOperation(ctx, state)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "op-123", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), v1beta1.GcpNfsBackupCreating, fromK8s.Status.State)
	assert.Equal(s.T(), 0, len(fromK8s.Status.Conditions))
}

func (s *checkBackupOperationSuite) TestCheckBackupOperationNil() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *file.Operation
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = v1beta1.GcpNfsBackupCreating
	obj.Status.OpIdentifier = "op-123"
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.Scope = &scope
	err, ctx = checkBackupOperation(ctx, state)
	assert.Equal(s.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), v1beta1.GcpNfsBackupError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (s *checkBackupOperationSuite) TestCheckBackupOperationCompletedFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opResp := file.Operation{
			Name: "op-123",
			Done: true,
			Error: &file.Status{
				Code:    500,
				Message: "internal error",
			},
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.OpIdentifier = "op-123"
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.Scope = &scope
	err, postCtx := checkBackupOperation(ctx, state)
	assert.Equal(s.T(), composed.StopAndForget, err)
	assert.Equal(s.T(), ctx, postCtx)
	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), v1beta1.GcpNfsBackupError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (s *checkBackupOperationSuite) TestCheckBackupOperationCompletedSucceeded() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opResp := file.Operation{
			Name: "op-123",
			Done: true,
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(s.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.OpIdentifier = "op-123"
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.Scope = &scope
	err, ctx = checkBackupOperation(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), v1beta1.GcpNfsBackupReady, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeReady, fromK8s.Status.Conditions[0].Type)
}

func TestCheckBackupOperation(t *testing.T) {
	suite.Run(t, new(checkBackupOperationSuite))
}
