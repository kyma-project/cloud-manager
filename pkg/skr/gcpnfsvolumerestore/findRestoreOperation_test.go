package gcpnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
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

type findRestoreOperationSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *findRestoreOperationSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *findRestoreOperationSuite) TestFindRestoreOperationRunningOp() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		op := file.Operation{
			Name: "op-123",
			Done: false,
		}
		opResp := file.ListOperationsResponse{
			Operations: []*file.Operation{&op},
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
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = scope.DeepCopy()
	err, ctx = findRestoreOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "op-123", fromK8s.Status.OpIdentifier)
}

func (suite *findRestoreOperationSuite) TestFindRestoreOperationNoRunningOp() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		op := file.Operation{
			Name: "op-123",
			Done: true,
		}
		opResp := file.ListOperationsResponse{
			Operations: []*file.Operation{&op},
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
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = scope.DeepCopy()

	err, ctx = findRestoreOperation(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", fromK8s.Status.OpIdentifier)
}

func (suite *findRestoreOperationSuite) TestFindRestoreOperationErrorNotFound() {
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
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = &scope
	err, ctx = findRestoreOperation(ctx, state)
	assert.Nil(suite.T(), err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), "", fromK8s.Status.State)
}

func (suite *findRestoreOperationSuite) TestFindRestoreOperationOtherError() {
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

	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = &scope
	err, ctx = findRestoreOperation(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(util.Timing.T100ms()), err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(suite.T(), v1beta1.JobStateError, fromK8s.Status.State)
	assert.Equal(suite.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func TestFindRestoreOperation(t *testing.T) {
	suite.Run(t, new(findRestoreOperationSuite))
}
