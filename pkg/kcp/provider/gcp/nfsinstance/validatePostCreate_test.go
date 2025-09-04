package nfsinstance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type validatePostCreateSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *validatePostCreateSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *validatePostCreateSuite) TestValidatePostCreateScaleDownFail() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name: "test-gcp-nfs-volume",
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: 5000,
			},
		},
	}
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = validatePostCreate(ctx, testState.State)
	assert.NotNil(s.T(), err)
	// validate status conditions error
	assert.Equal(s.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonValidationFailed, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonValidationFailed, updatedObject.Status.Conditions[0].Reason)
}

func (s *validatePostCreateSuite) TestValidatePostCreateSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstance()
	gcpNfsInstance.Spec.Instance.Gcp.Tier = v1beta1.ZONAL
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name: "test-gcp-nfs-volume",
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: 5000,
			},
		},
	}
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	err, resCtx := validatePostCreate(ctx, testState.State)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), resCtx)
	updatedObject2 := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject2)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), updatedObject.ResourceVersion, updatedObject2.ResourceVersion)
}

func TestValidatePostCreate(t *testing.T) {
	suite.Run(t, new(validatePostCreateSuite))
}
