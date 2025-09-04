package gcpnfsvolumerestore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type findRestoreOperationSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *findRestoreOperationSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *findRestoreOperationSuite) TestFindRestoreOperationRunningOp() {
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
			assert.Fail(s.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = scope.DeepCopy()
	err, ctx = findRestoreOperation(ctx, state)
	assert.Equal(s.T(), composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "op-123", fromK8s.Status.OpIdentifier)
}

func (s *findRestoreOperationSuite) TestFindRestoreOperationNoRunningOp() {
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
			assert.Fail(s.T(), "unable to marshal response: "+err.Error())
		}
		_, err = w.Write(b)
		if err != nil {
			assert.Fail(s.T(), "unable to write to provided ResponseWriter: "+err.Error())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = scope.DeepCopy()

	err, ctx = findRestoreOperation(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx, types.NamespacedName{Name: obj.Name, Namespace: obj.Namespace}, fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "", fromK8s.Status.OpIdentifier)
}

func (s *findRestoreOperationSuite) TestFindRestoreOperationErrorNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = &scope
	err, ctx = findRestoreOperation(ctx, state)
	assert.Nil(s.T(), err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), "", fromK8s.Status.State)
}

func (s *findRestoreOperationSuite) TestFindRestoreOperationOtherError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer fakeHttpServer.Close()

	//Get state object with GcpNfsVolume
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)

	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()
	state.Scope = &scope
	err, ctx = findRestoreOperation(ctx, state)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(util.Timing.T100ms()), err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "", fromK8s.Status.OpIdentifier)
	assert.Equal(s.T(), v1beta1.JobStateError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func TestFindRestoreOperation(t *testing.T) {
	suite.Run(t, new(findRestoreOperationSuite))
}
