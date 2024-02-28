package nfsinstance

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type validatePostCreateSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *validatePostCreateSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *validatePostCreateSuite) TestValidatePostCreateScaleDownFail() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

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
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := validatePostCreate(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.NotNil(suite.T(), err)
	// validate status conditions error
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonValidationFailed, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonValidationFailed, updatedObject.Status.Conditions[0].Reason)
}

func (suite *validatePostCreateSuite) TestValidatePostCreateSuccess() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstance()
	gcpNfsInstance.Spec.Instance.Gcp.Tier = v1beta1.ZONAL
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

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
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	err, resCtx := validatePostCreate(ctx, testState.State)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), resCtx)
	updatedObject2 := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject2)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), updatedObject.ObjectMeta.ResourceVersion, updatedObject2.ObjectMeta.ResourceVersion)
}

func TestValidatePostCreate(t *testing.T) {
	suite.Run(t, new(validatePostCreateSuite))
}
