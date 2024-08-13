package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type loadGcpNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *loadGcpNfsVolumeBackupSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 404
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolumeBackup.DeepCopy()

	factory, err := newTestStateFactoryWithObj(fakeHttpServer, objDiffName)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(objDiffName)
	suite.Nil(err)
	state.Scope = &scope

	//Invoke loadNfsBackup API
	err, _ctx := loadNfsBackup(ctx, state)

	//validate expected return values
	suite.Nil(err)
	suite.Nil(_ctx)
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupOtherError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 500
				http.Error(w, "Internal error", http.StatusInternalServerError)
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolumeBackup.DeepCopy()

	factory, err := newTestStateFactoryWithObj(fakeHttpServer, objDiffName)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(objDiffName)
	suite.Nil(err)
	state.Scope = &scope

	//Invoke loadNfsBackup API
	err, _ctx := loadNfsBackup(ctx, state)

	//validate expected return values
	suite.Equal(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), err)
	suite.Equal(ctx, _ctx)
}

func (suite *loadGcpNfsVolumeBackupSuite) TestVolumeBackupReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"test-gcp-nfs-volume-backup"}`))
			} else {
				assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.Scope = &scope

	//Invoke loadNfsBackup API
	err, ctx = loadNfsBackup(ctx, state)
	assert.Nil(suite.T(), err)
	assert.Nil(suite.T(), ctx)
}

func TestLoadGcpNfsVolumeBackupSuite(t *testing.T) {
	suite.Run(t, new(loadGcpNfsVolumeBackupSuite))
}
