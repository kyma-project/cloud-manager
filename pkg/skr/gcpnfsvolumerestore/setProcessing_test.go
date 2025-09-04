package gcpnfsvolumerestore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type setProcessingSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *setProcessingSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *setProcessingSuite) TestSetProcessingWhenDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := deletingGcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Nil(err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateDone() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateDone
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    v1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Equal(composed.StopAndForget, err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateFailed() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateFailed
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    v1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Equal(composed.StopAndForget, err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateInProgress() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateInProgress
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Nil(err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.JobStateError
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    v1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Nil(err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateEmpty() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeRestore.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Equal(composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolumeRestore{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal("", fromK8s.Status.OpIdentifier)
	s.Equal(v1beta1.JobStateProcessing, fromK8s.Status.State)
	s.Nil(fromK8s.Status.Conditions)
}

func TestSetProcessing(t *testing.T) {
	suite.Run(t, new(setProcessingSuite))
}
