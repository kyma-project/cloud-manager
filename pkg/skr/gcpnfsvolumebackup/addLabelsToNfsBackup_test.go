package gcpnfsvolumebackup

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	"io"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type addLabelsToNfsBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *addLabelsToNfsBackupSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *addLabelsToNfsBackupSuite) TestAddLabelsToNfsBackup() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPatch:
			fmt.Println(r.URL.Path)
			b, err := io.ReadAll(r.Body)
			assert.Nil(suite.T(), err)
			//create filestore instance from byte[] and check if it is equal to the expected filestore instance
			obj := &file.Backup{}
			err = json.Unmarshal(b, obj)
			suite.Nil(err)
			suite.Equal("cloud-manager", obj.Labels["managed-by"])
			suite.Equal(scope.Name, obj.Labels["scope-name"])
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	state.Scope = &scope
	assert.Nil(suite.T(), err)
	state.fileBackup = &file.Backup{}
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Equal(suite.T(), composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), err)
}

func (suite *addLabelsToNfsBackupSuite) TestDoNotAddLabelsToNfsBackupOnDeletingObject() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	deletingObj := deletingGpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, deletingObj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state, err := factory.newStateWith(deletingObj)
	assert.Nil(suite.T(), err)
	state.fileBackup = &file.Backup{}
	state.Scope = &scope

	//Call addLabelsToNfsBackup
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Nil(suite.T(), err)
}

func (suite *addLabelsToNfsBackupSuite) TestDoNotAddLabelsToNfsBackupNoBackup() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.Scope = &scope
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Nil(suite.T(), err)
}

func (suite *addLabelsToNfsBackupSuite) TestDoNotAddLabelsToNfsBackupNotReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	assert.Nil(suite.T(), err)
	state.fileBackup = &file.Backup{}
	state.Scope = &scope
	err, _ = addLabelsToNfsBackup(ctx, state)
	assert.Nil(suite.T(), err)
}

func TestAddLabelsToNfsBackup(t *testing.T) {
	suite.Run(t, new(addLabelsToNfsBackupSuite))
}
