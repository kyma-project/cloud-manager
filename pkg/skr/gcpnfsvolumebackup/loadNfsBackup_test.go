package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type loadGcpNfsVolumeBackupSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *loadGcpNfsVolumeBackupSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupNotFound() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
			fmt.Println(r.URL.Path)
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 404
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolumeBackup.DeepCopy()

	factory, err := newTestStateFactoryWithObj(fakeHttpServer, objDiffName)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(objDiffName)
	s.Nil(err)
	state.Scope = &scope

	//Invoke loadNfsBackup API
	err, _ctx := loadNfsBackup(ctx, state)

	//validate expected return values
	s.Nil(err)
	s.Nil(_ctx)
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupOtherError() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodGet:
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
	defer fakeHttpServer.Close()
	objDiffName := gcpNfsVolumeBackup.DeepCopy()

	factory, err := newTestStateFactoryWithObj(fakeHttpServer, objDiffName)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolumeBackup
	state, err := factory.newStateWith(objDiffName)
	s.Nil(err)
	state.Scope = &scope

	//Invoke loadNfsBackup API
	err, _ctx := loadNfsBackup(ctx, state)

	//validate expected return values
	s.Equal(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), err)
	s.Equal(ctx, _ctx)
}

func (s *loadGcpNfsVolumeBackupSuite) TestVolumeBackupReady() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.HasSuffix(r.URL.Path, "/projects/test-project/locations/us-west1/backups/cm-cffd6896-0127-48a1-8a64-e07f6ad5c912") {
				//Return 200
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"name":"test-gcp-nfs-volume-backup"}`))
			} else {
				assert.Fail(s.T(), "unexpected request: "+r.URL.String())
			}
		default:
			assert.Fail(s.T(), "unexpected request: "+r.URL.String())
		}
	}))
	defer fakeHttpServer.Close()
	obj := gcpNfsVolumeBackup.DeepCopy()
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//Get state object with GcpNfsVolume
	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.Scope = &scope

	//Invoke loadNfsBackup API
	err, ctx = loadNfsBackup(ctx, state)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), ctx)
}

func TestLoadGcpNfsVolumeBackupSuite(t *testing.T) {
	suite.Run(t, new(loadGcpNfsVolumeBackupSuite))
}
