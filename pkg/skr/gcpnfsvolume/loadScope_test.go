package gcpnfsvolume

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type loadScopeSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadScopeSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *loadScopeSuite) TestScopeNotFound() {
	factory, err := newTestStateFactory()
	s.Nil(err)

	//remove scope
	err = factory.kcpCluster.K8sClient().Delete(context.Background(), kcpScope.DeepCopy())
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	obj := gcpNfsVolume.DeepCopy()
	state := factory.newStateWith(obj)
	s.Nil(err)
	err, _ = loadScope(ctx, state)

	//validate expected return values
	s.Equal(composed.StopAndForget, err)
	fromK8s := &v1beta1.GcpNfsVolume{}

	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolume.Name,
			Namespace: gcpNfsVolume.Namespace},
		fromK8s)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.GcpNfsVolumeError, fromK8s.Status.State)
	assert.Equal(s.T(), metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	assert.Equal(s.T(), cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
}

func (s *loadScopeSuite) TestScopeExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()

	factory, err := newTestStateFactory()
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state := factory.newState()
	s.Nil(err)
	err, _ = loadScope(ctx, state)
	assert.Nil(s.T(), err)
}

func TestLoadScopeSuite(t *testing.T) {
	suite.Run(t, new(loadScopeSuite))
}
