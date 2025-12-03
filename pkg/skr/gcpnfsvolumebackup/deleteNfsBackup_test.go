package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type deleteNfsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *deleteNfsBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *deleteNfsBackupSuite) TestWhenBackupIsNotDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke deleteNfsBackup API
	state.fileBackup = &file.Backup{}
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *deleteNfsBackupSuite) TestWhenGcpBackupNotExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	state.fileBackup = nil
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *deleteNfsBackupSuite) TestWhenDeleteBackupReturnsError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodDelete:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 500
				http.Error(w, "Internal error", http.StatusInternalServerError)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = nil
	state.fileBackup = &file.Backup{}

	//Invoke deleteNfsBackup API
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime), err)
	s.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	_ = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletingGpNfsVolumeBackup.Name,
			Namespace: deletingGpNfsVolumeBackup.Namespace},
		fromK8s)

	//Validate expected status
	s.Equal(v1beta1.GcpNfsBackupError, fromK8s.Status.State)
	s.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	s.Equal(cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
	s.Equal(cloudcontrolv1beta1.ReasonGcpError, fromK8s.Status.Conditions[0].Reason)
}

func (s *deleteNfsBackupSuite) TestWhenDeleteBackupSuccessful() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodDelete:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"test-gcp-nfs-volume-backup-operation-id"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = nil
	state.fileBackup = &file.Backup{}

	//Invoke deleteNfsBackup API
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletingGpNfsVolumeBackup.Name,
			Namespace: deletingGpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err, "unexpected error")

	//Validate expected status
	s.Equal(v1beta1.GcpNfsBackupDeleting, fromK8s.Status.State)
	s.Equal(0, len(fromK8s.Status.Conditions))
}

func (s *deleteNfsBackupSuite) TestWhenDeleteBackupSuccessfulWithStatusLocation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodDelete:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"test-gcp-nfs-volume-backup-operation-id"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = nil
	state.fileBackup = &file.Backup{}

	//Invoke deleteNfsBackup API
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletingGpNfsVolumeBackup.Name,
			Namespace: deletingGpNfsVolumeBackup.Namespace},
		fromK8s)
	s.Nil(err, "unexpected error")

	//Validate expected status
	s.Equal(v1beta1.GcpNfsBackupDeleting, fromK8s.Status.State)
	s.Equal(0, len(fromK8s.Status.Conditions))
}

func (s *deleteNfsBackupSuite) TestWhenDeleteBackupReturns403() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodDelete:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 403 (invalid location)
				http.Error(w, "Location us-west15-a is not found or access is unauthorized.", http.StatusForbidden)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = nil
	state.fileBackup = &file.Backup{}

	//Invoke deleteNfsBackup API
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values - 403 should be treated as success (resource not found)
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *deleteNfsBackupSuite) TestWhenDeleteBackupReturns404() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodDelete:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 404 (not found)
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = nil
	state.fileBackup = &file.Backup{}

	//Invoke deleteNfsBackup API
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values - 404 should be treated as success (resource not found)
	s.Nil(err)
	s.Nil(_ctx)
}

func TestDeleteNfsBackup(t *testing.T) {
	suite.Run(t, new(deleteNfsBackupSuite))
}
