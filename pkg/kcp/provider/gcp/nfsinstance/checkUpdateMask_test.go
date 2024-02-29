package nfsinstance

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

type checkUpdateMaskSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *checkUpdateMaskSuite) SetupTest() {
	suite.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (suite *checkUpdateMaskSuite) TestCheckUpdateMaskModifyIncrease() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "test-gcp-nfs-volume-2",
		State: string(client.READY),
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: 500,
			},
		},
	}
	testState.operation = client.MODIFY
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "FileShares", testState.State.updateMask[0])
}

func (suite *checkUpdateMaskSuite) TestCheckUpdateMaskModifyDecrease() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "test-gcp-nfs-volume-2",
		State: string(client.READY),
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: 5000,
			},
		},
	}
	testState.operation = client.MODIFY
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "FileShares", testState.State.updateMask[0])
}

func (suite *checkUpdateMaskSuite) TestCheckUpdateMaskModifyNoChange() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "test-gcp-nfs-volume-2",
		State: string(client.READY),
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: int64(gcpNfsInstance.Spec.Instance.Gcp.CapacityGb),
			},
		},
	}
	testState.operation = client.MODIFY
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(testState.State.updateMask))
}

func (suite *checkUpdateMaskSuite) TestCheckUpdateMaskNotModify() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(suite.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(suite.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Get state object with GcpNfsVolume

	testState, err := factory.newStateWith(ctx, gcpNfsInstance, "")
	testState.fsInstance = &file.Instance{
		Name:  "test-gcp-nfs-volume-2",
		State: string(client.READY),
		FileShares: []*file.FileShareConfig{
			{
				CapacityGb: 500,
			},
		},
	}
	//NONE operation
	testState.operation = client.NONE
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(testState.State.updateMask))

	//ADD operation
	testState.operation = client.ADD
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx = checkUpdateMask(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(testState.State.updateMask))

	//DELETE operation
	testState.operation = client.DELETE
	assert.Nil(suite.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx = checkUpdateMask(ctx, testState.State)
	assert.Nil(suite.T(), resCtx)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(testState.State.updateMask))
}

func TestCheckUpdateMask(t *testing.T) {
	suite.Run(t, new(checkUpdateMaskSuite))
}
