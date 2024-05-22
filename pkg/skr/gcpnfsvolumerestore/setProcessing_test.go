package gcpnfsvolumerestore

import (
	"context"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type setProcessingSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *setProcessingSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *setProcessingSuite) TestSetProcessingWhenDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := deletingGpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateDone() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateDone
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Equal(composed.StopAndForget, err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateFailed
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Equal(composed.StopAndForget, err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateInProgress() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateInProgress
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateError
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    cloudresourcesv1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Nil(err)
	suite.Nil(ctx)
}

func (suite *setProcessingSuite) TestSetProcessingWhenStateEmpty() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, ctx = setProcessing(ctx, state)
	suite.Equal(composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	suite.Nil(err)
	suite.Equal("", fromK8s.Status.OpIdentifier)
	suite.Equal(v1beta1.JobStateProcessing, fromK8s.Status.State)
	suite.Nil(fromK8s.Status.Conditions)
}

func TestSetProcessing(t *testing.T) {
	suite.Run(t, new(setProcessingSuite))
}
