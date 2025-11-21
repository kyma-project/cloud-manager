package gcpnfsvolume

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
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

	obj := deletedGcpNfsVolume.DeepCopy()
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)

	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Nil(err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateDone() {

	obj := gcpNfsVolume.DeepCopy()
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.StateReady
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Nil(err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateError() {

	obj := gcpNfsVolume.DeepCopy()
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	obj.Status.State = v1beta1.StateError
	meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  "test",
		Message: "test",
	})
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Nil(err)
	s.Nil(ctx)
}

func (s *setProcessingSuite) TestSetProcessingWhenStateEmpty() {

	obj := gcpNfsVolume.DeepCopy()
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	factory, err := newTestStateFactory(fakeHttpServer)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Set the Status.State to empty.
	obj.Status.State = ""

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	s.Nil(err)

	err, ctx = setProcessing(ctx, state)
	s.Equal(composed.StopWithRequeue, err)
	fromK8s := &v1beta1.GcpNfsVolume{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: obj.Name,
			Namespace: obj.Namespace},
		fromK8s)
	s.Nil(err)
	s.Equal(v1beta1.GcpNfsVolumeProcessing, fromK8s.Status.State)
	s.Nil(fromK8s.Status.Conditions)
}

func TestSetProcessing(t *testing.T) {
	suite.Run(t, new(setProcessingSuite))
}
