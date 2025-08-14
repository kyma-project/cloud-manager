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

func (suite *updateCapacitySuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *updateCapacitySuite) TestWhileDeleting() {

	obj := deletingGpNfsVolumeBackup.DeepCopy()

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)

	err, _ctx := updateCapacity(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)

}

func (suite *updateCapacitySuite) TestLastCapacityUpdateNil() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.LastCapacityUpdate = nil

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.fileBackup = &file.Backup{
		CapacityGb: 1024,
	}
	suite.NotNil(state.fileBackup)

	err, _ctx := updateCapacity(ctx, state)
	suite.Nil(err)
	suite.NotNil(_ctx)

}

func (suite *updateCapacitySuite) TestLastCapacityUpdateIsZero() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	var timeZero time.Time
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: timeZero}

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.fileBackup = &file.Backup{
		CapacityGb: 1024,
	}
	suite.NotNil(state.fileBackup)

	err, _ctx := updateCapacity(ctx, state)
	suite.Nil(err)
	suite.NotNil(_ctx)

}

func (suite *updateCapacitySuite) TestLastCapacityUpdateGreater() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().Add(-5 * time.Second)}

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.fileBackup = &file.Backup{
		CapacityGb: 1024,
	}
	suite.NotNil(state.fileBackup)

	gcpclient.GcpConfig.GcpCapacityCheckInterval = time.Second * 5
	// gcpclient.GcpConfig.GcpCapacityCheckInterval = time.Hour * 1

	err, _ctx := updateCapacity(ctx, state)
	suite.Nil(err)
	suite.NotNil(_ctx)

}

func (suite *updateCapacitySuite) TestLastCapacityUpdateLesser() {

	obj := gcpNfsVolumeBackup.DeepCopy()
	obj.Status.LastCapacityUpdate = &metav1.Time{Time: time.Now().Add(-1 * time.Second)}

	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	factory, err := newTestStateFactoryWithObj(fakeHttpServer, obj)
	suite.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	state, err := factory.newStateWith(obj)
	suite.Nil(err)
	state.fileBackup = &file.Backup{
		CapacityGb: 1024,
	}
	suite.NotNil(state.fileBackup)

	gcpclient.GcpConfig.GcpCapacityCheckInterval = time.Hour * 1

	err, _ctx := updateCapacity(ctx, state)
	suite.Nil(err)
	suite.Nil(_ctx)

}

func TestUpdateCapacityGCP(t *testing.T) {
	suite.Run(t, new(updateCapacitySuite))
}
