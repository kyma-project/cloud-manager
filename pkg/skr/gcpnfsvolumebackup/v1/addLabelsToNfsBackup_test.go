package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type addLabelsToNfsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *addLabelsToNfsBackupSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *addLabelsToNfsBackupSuite) TestAddLabelsToNfsBackup() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPatch:
			fmt.Println(r.URL.Path)
			b, err := io.ReadAll(r.Body)
			assert.Nil(s.T(), err)
			//create filestore instance from byte[] and check if it is equal to the expected filestore instance
			obj := &file.Backup{}
			err = json.Unmarshal(b, obj)
			s.Nil(err)
			s.Equal("cloud-manager", obj.Labels["managed-by"])
			s.Equal(scope.Name, obj.Labels["scope-name"])
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	state.Scope = &scope
	assert.Nil(s.T(), err)
	state.fileBackup = &file.Backup{}
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Equal(s.T(), composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime), err)
}

func (s *addLabelsToNfsBackupSuite) TestDoNotAddLabelsToNfsBackupOnDeletingObject() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	deletingObj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, deletingObj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	assert.Nil(s.T(), err)
	state.fileBackup = &file.Backup{}
	state.Scope = &scope

	//Call addLabelsToNfsBackup
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Nil(s.T(), err)
}

func (s *addLabelsToNfsBackupSuite) TestDoNotAddLabelsToNfsBackupNoBackup() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.Scope = &scope
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Nil(s.T(), err)
}

func (s *addLabelsToNfsBackupSuite) TestDoNotAddLabelsToNfsBackupNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(s.T(), err)
	state.fileBackup = &file.Backup{}
	state.Scope = &scope
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Nil(s.T(), err)
}

func TestAddLabelsToNfsBackup(t *testing.T) {
	suite.Run(t, new(addLabelsToNfsBackupSuite))
}
