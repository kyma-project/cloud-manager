package gcpnfsvolumebackup

import (
	"context"
	"net/http"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/go-logr/logr"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type updateCapacitySuite struct {
	suite.Suite
	ctx context.Context
}

func (s *updateCapacitySuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *updateCapacitySuite) TestWhileDeleting() {

	obj := deletingGpNfsVolumeBackup.DeepCopy()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	s.Nil(err)

	err, _ctx := updateCapacity(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)

}

func (s *updateCapacitySuite) TestLastCapacityUpdateNil() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.LastCapacityUpdate = nil

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.fileBackup = &file.Backup{
		StorageBytes: 1024,
	}
	s.NotNil(state.fileBackup)

	err, _ctx := updateCapacity(ctx, state)
	s.Nil(err)
	s.NotNil(_ctx)

}

func (s *updateCapacitySuite) TestLastCapacityUpdateIsZero() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	var timeZero time.Time
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: timeZero}

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.fileBackup = &file.Backup{
		StorageBytes: 1024,
	}
	s.NotNil(state.fileBackup)

	err, _ctx := updateCapacity(ctx, state)
	s.Nil(err)
	s.NotNil(_ctx)

}

func (s *updateCapacitySuite) TestLastCapacityUpdateGreater() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().Add(-5 * time.Second)}

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.fileBackup = &file.Backup{
		StorageBytes: 1024,
	}
	s.NotNil(state.fileBackup)

	gcpclient.GcpConfig.GcpCapacityCheckInterval = time.Second * 5
	// gcpclient.GcpConfig.GcpCapacityCheckInterval = time.Hour * 1

	err, _ctx := updateCapacity(ctx, state)
	s.Nil(err)
	s.NotNil(_ctx)

}

func (s *updateCapacitySuite) TestLastCapacityUpdateLesser() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().Add(-1 * time.Second)}

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	s.Nil(err)
	state.fileBackup = &file.Backup{
		StorageBytes: 1024,
	}
	s.NotNil(state.fileBackup)

	gcpclient.GcpConfig.GcpCapacityCheckInterval = time.Hour * 1

	err, _ctx := updateCapacity(ctx, state)
	s.Nil(err)
	s.Nil(_ctx)

}

func TestUpdateCapacityGCP(t *testing.T) {
	suite.Run(t, new(updateCapacitySuite))
}
