package gcpnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type checkRestoreOperationSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *checkRestoreOperationSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *checkRestoreOperationSuite) TestCheckRestoreOperationNoOpId() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	err, ctx = checkRestoreOperation(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func (suite *checkRestoreOperationSuite) TestCheckRestoreOperationErrorNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.OpIdentifier = "op-123"
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.Scope = &scope
	err, ctx = checkRestoreOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), v1beta1.StateError, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (suite *checkRestoreOperationSuite) TestCheckRestoreOperationOtherError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer fakeHttpServer.Close()

	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj.Status.OpIdentifier = "op-123"
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.Scope = &scope
	err, ctx = checkRestoreOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "op-123", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), v1beta1.StateError, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (suite *checkRestoreOperationSuite) TestCheckRestoreOperationNotCompleted() {
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
	}))
	defer fakeHttpServer.Close()
	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeRestore.DeepCopy()
	obj.Status.OpIdentifier = "op-123"
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.Scope = &scope
	err, ctx = checkRestoreOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "op-123", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), "", fromK8s.Status.State)
	assert.Equal(suite.T(), 0, len(fromK8s.Status.Conditions))
}

func (suite *checkRestoreOperationSuite) TestCheckRestoreOperationNil() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opResp *file.Operation
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(suite.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	obj.Status.OpIdentifier = "op-123"
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.Scope = &scope
	err, ctx = checkRestoreOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), v1beta1.StateError, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (suite *checkRestoreOperationSuite) TestCheckRestoreOperationCompletedFailed() {
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
			assert.Fail(suite.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeRestore.DeepCopy()
	obj.Status.OpIdentifier = "op-123"
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.Scope = &scope
	err, ctx = checkRestoreOperation(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), v1beta1.JobStateFailed, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (suite *checkRestoreOperationSuite) TestCheckRestoreOperationCompletedSucceeded() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opResp := file.Operation{
			Name: "op-123",
			Done: true,
		}
		b, err := json.Marshal(opResp)
		if err != nil {
			assert.Fail(suite.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(suite.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
		return
	}))
	defer fakeHttpServer.Close()
	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeRestore.DeepCopy()
	obj.Status.OpIdentifier = "op-123"
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.Scope = &scope
	err, ctx = checkRestoreOperation(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), v1beta1.JobStateDone, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), cloudcontrolv1beta1.ConditionTypeReady, fromK8s.Status.Conditions[0].Type)
}

func TestCheckRestoreOperation(t *testing.T) {
	suite.Run(t, new(checkRestoreOperationSuite))
}
