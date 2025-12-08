package gcpnfsvolumebackup

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type createNfsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *createNfsBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *createNfsBackupSuite) TestWhenBackupIsDeleting() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	obj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Invoke createNfsBackup API
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *createNfsBackupSuite) TestWhenGcpBackupExists() {
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

	state.fileBackup = &file.Backup{}
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *createNfsBackupSuite) TestWhenCreateBackupReturnsError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPost:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups") {
				//Return 500
				http.Error(w, "Internal error", http.StatusInternalServerError)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()

	//Invoke createNfsBackup API
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
	s.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	_ = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)

	//Validate expected status
	s.Equal(v1beta1.GcpNfsBackupError, fromK8s.Status.State)
	s.Equal(metav1.ConditionTrue, fromK8s.Status.Conditions[0].Status)
	s.Equal(cloudcontrolv1beta1.ConditionTypeError, fromK8s.Status.Conditions[0].Type)
	s.Equal(cloudcontrolv1beta1.ReasonGcpError, fromK8s.Status.Conditions[0].Reason)
}

func (s *createNfsBackupSuite) TestWhenCreateBackupNoIdNoLocation() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPost:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups") {
				//Return 500
				http.Error(w, "Internal error", http.StatusInternalServerError)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.Id = ""
	obj.Status.Location = ""
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()

	//Invoke createNfsBackup API
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeue, err)
	s.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	_ = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)

	//Validate expected status
	s.Equal("us-west1", fromK8s.Status.Location)
	s.NotEqual("", fromK8s.Status.Id)
}

func (s *createNfsBackupSuite) TestWhenCreateBackupSuccessful() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPost:
			fmt.Println(r.URL.Path)
			b, err := io.ReadAll(r.Body)
			assert.Nil(s.T(), err)
			//create filestore instance from byte[] and check if it is equal to the expected filestore instance
			obj := &file.Backup{}
			err = json.Unmarshal(b, obj)
			s.Nil(err)
			s.Equal("projects/test-project/locations/us-west1/instances/cm-test-gcp-nfs-instance", obj.SourceInstance)
			s.Equal("cloud-manager", obj.Labels["managed-by"])
			s.Equal(scope.Name, obj.Labels["scope-name"])
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups") && strings.Contains(r.URL.RawQuery, "backupId=cm-") {
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
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)

	//Set the scope and gcpNfsVolume objects in state
	state.Scope = scope.DeepCopy()
	state.GcpNfsVolume = gcpNfsVolume.DeepCopy()

	//Invoke createNfsBackup API
	err, _ctx := createNfsBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
	s.NotNil(_ctx)

	fromK8s := &v1beta1.GcpNfsVolumeBackup{}
	_ = factory.skrCluster.K8sClient().Get(ctx,
		types.NamespacedName{Name: gcpNfsVolumeBackup.Name,
			Namespace: gcpNfsVolumeBackup.Namespace},
		fromK8s)

	//Validate expected status
	s.Equal(v1beta1.GcpNfsBackupCreating, fromK8s.Status.State)
	s.Equal(0, len(fromK8s.Status.Conditions))
}

func TestCreateNfsBackup(t *testing.T) {
	suite.Run(t, new(createNfsBackupSuite))
}
