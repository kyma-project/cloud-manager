package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"testing"
)

type deleteNfsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *deleteNfsBackupSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *deleteNfsBackupSuite) TestWhenBackupIsNotDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Invoke deleteNfsBackup API
	state.fileBackup = &file.Backup{}
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteNfsBackupSuite) TestWhenGcpBackupNotExists() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	state.fileBackup = nil
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *deleteNfsBackupSuite) TestWhenDeleteBackupReturnsError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodDelete:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/test-gcp-nfs-volume-restore") {
				//Return 500
				http.Error(w, "Internal error", http.StatusInternalServerError)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = nil
	state.fileBackup = &file.Backup{}

	//Invoke deleteNfsBackup API
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
	suite.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	_ = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletingGpNfsVolumeBackup.Name,
			Namespace: deletingGpNfsVolumeBackup.Namespace},
		fromK8s)

	//Validate expected status
	suite.Equal(v1beta1.GcpNfsBackupError, fromK8s.Status.State)
	suite.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	suite.Equal(cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
	suite.Equal(cloudcontrolv1beta1.ReasonGcpError, fromK8s.Status.Conditions[0].Reason)
}

func (suite *deleteNfsBackupSuite) TestWhenDeleteBackupSuccessful() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodDelete:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/test-gcp-nfs-volume-restore") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"test-gcp-nfs-volume-backup-operation-id"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = nil
	state.fileBackup = &file.Backup{}

	//Invoke deleteNfsBackup API
	err, _ctx := deleteNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	err = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: deletingGpNfsVolumeBackup.Name,
			Namespace: deletingGpNfsVolumeBackup.Namespace},
		fromK8s)
	suite.Nil(err, "unexpected error")

	//Validate expected status
	suite.Equal(v1beta1.GcpNfsBackupDeleting, fromK8s.Status.State)
	suite.Equal(0, len(fromK8s.Status.Conditions))
}

func TestDeleteNfsBackup(t *testing.T) {
	suite.Run(t, new(deleteNfsBackupSuite))
}
