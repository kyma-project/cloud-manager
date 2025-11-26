package nfsinstance

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type syncNfsInstanceSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *syncNfsInstanceSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *syncNfsInstanceSuite) TestSyncNfsInstanceAddSuccess() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(s.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(s.T(), err)
				assert.Equal(s.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"create-instance-operation"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = gcpclient.ADD
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), err)
	assert.NotNil(s.T(), testState.ObjAsNfsInstance().Status.Id)
	assert.Equal(s.T(), "create-instance-operation", testState.ObjAsNfsInstance().Status.OpIdentifier)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "create-instance-operation", updatedObject.Status.OpIdentifier)
}

func (s *syncNfsInstanceSuite) TestSyncNfsInstanceAddError() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(s.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(s.T(), err)
				assert.Equal(s.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"error"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = gcpclient.ADD
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	assert.Equal(s.T(), "", testState.ObjAsNfsInstance().Status.Id)
	// check conditions
	assert.Len(s.T(), testState.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, testState.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, testState.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, testState.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.NotEqual(s.T(), "", testState.ObjAsNfsInstance().Status.Conditions[0].Message)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, updatedObject.Status.Conditions[0].Reason)
	assert.NotEqual(s.T(), "", updatedObject.Status.Conditions[0].Message)
}

func (s *syncNfsInstanceSuite) TestSyncNfsInstancePatchSuccess() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(s.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(s.T(), err)
				assert.Equal(s.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"patch-instance-operation"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = gcpclient.MODIFY
	testState.updateMask = []string{"description"}
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), err)
	assert.NotNil(s.T(), testState.ObjAsNfsInstance().Status.Id)
	assert.Equal(s.T(), "patch-instance-operation", testState.ObjAsNfsInstance().Status.OpIdentifier)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "patch-instance-operation", updatedObject.Status.OpIdentifier)
}

func (s *syncNfsInstanceSuite) TestSyncNfsInstancePatchError() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				b, err := io.ReadAll(r.Body)
				assert.Nil(s.T(), err)
				//create filestore instance from byte[] and check if it is equal to the expected filestore instance
				obj := &file.Instance{}
				err = json.Unmarshal(b, obj)
				assert.Nil(s.T(), err)
				assert.Equal(s.T(), gcpNfsInstance.Name, obj.Description)
				//Return 200
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"error"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = gcpclient.MODIFY
	testState.updateMask = []string{"description"}
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	assert.Equal(s.T(), "", testState.ObjAsNfsInstance().Status.Id)
	// check conditions
	assert.Len(s.T(), testState.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, testState.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, testState.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, testState.ObjAsNfsInstance().Status.Conditions[0].Reason)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, updatedObject.Status.Conditions[0].Reason)
	assert.NotEqual(s.T(), "", updatedObject.Status.Conditions[0].Message)
}

func (s *syncNfsInstanceSuite) TestSyncNfsInstanceDeleteSuccess() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"delete-instance-operation"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = gcpclient.DELETE
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), err)
	assert.Equal(s.T(), "delete-instance-operation", testState.ObjAsNfsInstance().Status.OpIdentifier)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "delete-instance-operation", updatedObject.Status.OpIdentifier)
}

func (s *syncNfsInstanceSuite) TestSyncNfsInstanceDeleteError() {
	gcpNfsInstance := getGcpNfsInstance()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance") {
				//Return 200
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"error"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	testState.operation = gcpclient.DELETE
	err, _ = syncNfsInstance(ctx, testState.State)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	// check conditions
	assert.Len(s.T(), testState.ObjAsNfsInstance().Status.Conditions, 1)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, testState.ObjAsNfsInstance().Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, testState.ObjAsNfsInstance().Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, testState.ObjAsNfsInstance().Status.Conditions[0].Reason)
	assert.NotEqual(s.T(), "", testState.ObjAsNfsInstance().Status.Conditions[0].Message)

	updatedObject := &v1beta1.NfsInstance{}
	err = factory.kcpCluster.K8sClient().Get(ctx, types.NamespacedName{Name: gcpNfsInstance.Name, Namespace: gcpNfsInstance.Namespace}, updatedObject)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), v1beta1.ConditionTypeError, updatedObject.Status.Conditions[0].Type)
	assert.Equal(s.T(), metav1.ConditionTrue, updatedObject.Status.Conditions[0].Status)
	assert.Equal(s.T(), v1beta1.ReasonGcpError, updatedObject.Status.Conditions[0].Reason)
	assert.NotEqual(s.T(), "", updatedObject.Status.Conditions[0].Message)
}

func (s *syncNfsInstanceSuite) TestGetInstanceWithProtocol_BasicHdd() {
	gcpNfsInstance := getGcpNfsInstance()
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request.")
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	instance := getInstanceWithProtocol(ctx, testState.State)
	assert.Empty(s.T(), instance.Protocol)
	assert.Equal(s.T(), testState.toInstance(), instance)
}

func (s *syncNfsInstanceSuite) TestGetInstanceWithProtocol_BasicSsd() {
	gcpNfsInstance := getGcpNfsInstance()
	gcpNfsInstance.Spec.Instance.Gcp.Tier = v1beta1.BASIC_SSD
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request.")
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	instance := getInstanceWithProtocol(ctx, testState.State)
	assert.Empty(s.T(), instance.Protocol)
	assert.Equal(s.T(), testState.toInstance(), instance)
}

func (s *syncNfsInstanceSuite) TestGetInstanceWithProtocol_Zonal_Without_FF() {
	gcpNfsInstance := getGcpNfsInstance()
	gcpNfsInstance.Spec.Instance.Gcp.Tier = v1beta1.ZONAL
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request.")
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	instance := getInstanceWithProtocol(ctx, testState.State)
	assert.Empty(s.T(), instance.Protocol)
	assert.Equal(s.T(), testState.toInstance(), instance)
}

func (s *syncNfsInstanceSuite) TestGetInstanceWithProtocol_Zonal_With_FF() {

	gcpNfsInstance := getGcpNfsInstance()
	gcpNfsInstance.Spec.Instance.Gcp.Tier = v1beta1.ZONAL
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request.")
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = feature.Initialize(ctx, logr.Discard(), feature.WithFile("testdata/nfs41Enabled.yaml"))
	assert.NoError(s.T(), err)

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	instance := getInstanceWithProtocol(ctx, testState.State)
	assert.Equal(s.T(), string(gcpclient.FilestoreProtocolNFSv41), instance.Protocol)
	instance.Protocol = ""
	assert.Equal(s.T(), testState.toInstance(), instance)
	_ = feature.Initialize(ctx, logr.Discard())
}

func (s *syncNfsInstanceSuite) TestGetInstanceWithProtocol_Regional_Without_FF() {
	gcpNfsInstance := getGcpNfsInstance()
	gcpNfsInstance.Spec.Instance.Gcp.Tier = v1beta1.REGIONAL
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request.")
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	instance := getInstanceWithProtocol(ctx, testState.State)
	assert.Empty(s.T(), instance.Protocol)
	assert.Equal(s.T(), testState.toInstance(), instance)
}

func (s *syncNfsInstanceSuite) TestGetInstanceWithProtocol_Regional_With_FF() {
	gcpNfsInstance := getGcpNfsInstance()
	gcpNfsInstance.Spec.Instance.Gcp.Tier = v1beta1.REGIONAL
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request.")
	}))
	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = feature.Initialize(ctx, logr.Discard(), feature.WithFile("testdata/nfs41Enabled.yaml"))
	assert.NoError(s.T(), err)
	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	instance := getInstanceWithProtocol(ctx, testState.State)
	assert.Equal(s.T(), string(gcpclient.FilestoreProtocolNFSv41), instance.Protocol)
	instance.Protocol = ""
	assert.Equal(s.T(), testState.toInstance(), instance)
	_ = feature.Initialize(ctx, logr.Discard())
}

func TestSyncNfsInstance(t *testing.T) {
	suite.Run(t, new(syncNfsInstanceSuite))
}
