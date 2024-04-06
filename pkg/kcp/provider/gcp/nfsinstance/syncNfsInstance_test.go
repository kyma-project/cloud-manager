package nfsinstance

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"testing"
)

type syncNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *syncNfsInstanceSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *syncNfsInstanceSuite) TestSyncNfsInstanceAddSuccess() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(suite.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(suite.T(), err)
				assert.Equal(suite.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"create-instance-operation"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = client.ADD
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(testState.gcpConfig.GcpOperationWaitTime), err)
	assert.NotNil(suite.T(), testState.State.ObjAsNfsInstance().Status.Id)
	assert.Equal(suite.T(), "create-instance-operation", testState.State.ObjAsNfsInstance().Status.OpIdentifier)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "create-instance-operation", updatedObject.Status.OpIdentifier)
}

func (suite *syncNfsInstanceSuite) TestSyncNfsInstanceAddError() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(suite.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(suite.T(), err)
				assert.Equal(suite.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"error"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = client.ADD
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(testState.gcpConfig.GcpRetryWaitTime), err)
	assert.Equal(suite.T(), "", testState.State.ObjAsNfsInstance().Status.Id)
	// check conditions
	assert.Len(suite.T(), testState.State.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.NotEqual(suite.T(), "", testState.State.ObjAsNfsInstance().Status.Conditions[0].Message)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, updatedObject.Status.Conditions[0].Reason)
	assert.NotEqual(suite.T(), "", updatedObject.Status.Conditions[0].Message)
}

func (suite *syncNfsInstanceSuite) TestSyncNfsInstancePatchSuccess() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/test-gcp-nfs-volume") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(suite.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(suite.T(), err)
				assert.Equal(suite.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"patch-instance-operation"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = client.MODIFY
	testState.updateMask = []string{"description"}
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(testState.gcpConfig.GcpOperationWaitTime), err)
	assert.NotNil(suite.T(), testState.State.ObjAsNfsInstance().Status.Id)
	assert.Equal(suite.T(), "patch-instance-operation", testState.State.ObjAsNfsInstance().Status.OpIdentifier)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "patch-instance-operation", updatedObject.Status.OpIdentifier)
}

func (suite *syncNfsInstanceSuite) TestSyncNfsInstancePatchError() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/test-gcp-nfs-volume") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(suite.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(suite.T(), err)
				assert.Equal(suite.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"error"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = client.MODIFY
	testState.updateMask = []string{"description"}
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(testState.gcpConfig.GcpRetryWaitTime), err)
	assert.Equal(suite.T(), "", testState.State.ObjAsNfsInstance().Status.Id)
	// check conditions
	assert.Len(suite.T(), testState.State.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, updatedObject.Status.Conditions[0].Reason)
	assert.NotEqual(suite.T(), "", updatedObject.Status.Conditions[0].Message)
}

func (suite *syncNfsInstanceSuite) TestSyncNfsInstanceDeleteSuccess() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/test-gcp-nfs-volume") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"delete-instance-operation"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = client.DELETE
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(testState.gcpConfig.GcpOperationWaitTime), err)
	assert.Equal(suite.T(), "delete-instance-operation", testState.State.ObjAsNfsInstance().Status.OpIdentifier)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "delete-instance-operation", updatedObject.Status.OpIdentifier)
}

func (suite *syncNfsInstanceSuite) TestSyncNfsInstanceDeleteError() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/test-gcp-nfs-volume") {
				//Return 200
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"error"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = client.DELETE
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(testState.gcpConfig.GcpRetryWaitTime), err)
	// check conditions
	assert.Len(suite.T(), testState.State.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, testState.State.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, testState.State.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.NotEqual(suite.T(), "", testState.State.ObjAsNfsInstance().Status.Conditions[0].Message)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(suite.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(suite.T(), v1beta1.ReasonGcpError, updatedObject.Status.Conditions[0].Reason)
	assert.NotEqual(suite.T(), "", updatedObject.Status.Conditions[0].Message)
}

func TestSyncNfsInstance(t *testing.T) {
	suite.Run(t, new(syncNfsInstanceSuite))
}
