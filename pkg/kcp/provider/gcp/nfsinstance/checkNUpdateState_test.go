package nfsinstance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type checkNUpdateStateSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *checkNUpdateStateSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateStateDeleted() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getDeletedGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = nil
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkNUpdateState(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	assert.Equal(suite.T(), client.Deleted, testState.ObjAsNfsInstance().Status.State)
	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), client.Deleted, updatedObject.Status.State)
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateStateNotDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getDeletedGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name: "deleted-gcp-nfs-volume",
	}
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkNUpdateState(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	assert.Equal(suite.T(), client.DELETE, testState.operation)
	assert.Equal(suite.T(), client.Deleting, testState.curState)
	assert.Equal(suite.T(), client.Deleting, testState.ObjAsNfsInstance().Status.State)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), client.Deleting, updatedObject.Status.State)
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateStateDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getDeletedGcpNfsInstance()
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "deleted-gcp-nfs-volume",
		State: string(client.DELETING),
	}
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkNUpdateState(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Nil(suite.T(), err)
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateStateCreate() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = nil
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkNUpdateState(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	assert.Equal(suite.T(), client.ADD, testState.operation)
	assert.Equal(suite.T(), client.Creating, testState.curState)
	assert.Equal(suite.T(), client.Creating, testState.ObjAsNfsInstance().Status.State)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), client.Creating, updatedObject.Status.State)
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateStateUpdate() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "test-gcp-nfs-volume-2",
		State: string(client.READY),
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: 5000,
			},
		},
	}
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkNUpdateState(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), composed.StopWithRequeue, err)
	assert.Equal(suite.T(), client.MODIFY, testState.operation)
	assert.Equal(suite.T(), client.Updating, testState.curState)
	assert.Equal(suite.T(), client.Updating, testState.ObjAsNfsInstance().Status.State)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), client.Updating, updatedObject.Status.State)
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateStateReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "test-gcp-nfs-volume-2",
		State: string(client.READY),
		Networks: []*file.NetworkConfig{
			{
				IpAddresses: []string{"10.2.74.33"},
			},
		},
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: int64(gcpNfsInstance.Spec.Instance.Gcp.CapacityGb),
			},
		},
	}
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkNUpdateState(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), composed.StopAndForget, err)
	assert.Equal(suite.T(), v1beta1.StateReady, testState.curState)
	assert.Equal(suite.T(), v1beta1.StateReady, testState.ObjAsNfsInstance().Status.State)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.StateReady, updatedObject.Status.State)
	// validate status conditions
	assert.Equal(suite.T(), v1beta1.ConditionTypeReady, updatedObject.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonReady, updatedObject.Status.Conditions[0].Reason)
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateStateError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:          "test-gcp-nfs-volume-2",
		State:         string(client.ERROR),
		StatusMessage: "Some error",
	}
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkNUpdateState(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(client.GcpConfig.GcpRetryWaitTime), err)
	assert.Equal(suite.T(), v1beta1.StateError, testState.curState)
	assert.Equal(suite.T(), v1beta1.StateError, testState.ObjAsNfsInstance().Status.State)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.StateError, updatedObject.Status.State)
	// validate status conditions
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, updatedObject.Status.Conditions[0].Reason)
}

func (suite *checkNUpdateStateSuite) TestCheckNUpdateFilestoreStateTransient() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "test-gcp-nfs-volume-2",
		State: "somethingElse",
		Networks: []*file.NetworkConfig{
			{
				IpAddresses: []string{"10.2.74.33"},
			},
		},
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: int64(gcpNfsInstance.Spec.Instance.Gcp.CapacityGb),
			},
		},
	}
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, _ = checkNUpdateState(ctx, testState.State)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(client.GcpConfig.GcpRetryWaitTime), err)
	assert.Equal(suite.T(), gcpNfsInstance.Status.State, testState.curState)
	assert.Equal(suite.T(), client.NONE, testState.operation)
}

func TestCheckNUpdateState(t *testing.T) {
	suite.Run(t, new(checkNUpdateStateSuite))
}
