package nfsinstance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/file/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type checkUpdateMaskSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *checkUpdateMaskSuite) SetupTest() {
	s.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (s *checkUpdateMaskSuite) TestCheckUpdateMaskModifyIncrease() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(s.T(), resCtx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "FileShares", testState.updateMask[0])
}

func (s *checkUpdateMaskSuite) TestCheckUpdateMaskModifyDecrease() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(s.T(), resCtx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "FileShares", testState.updateMask[0])
}

func (s *checkUpdateMaskSuite) TestCheckUpdateMaskModifyNoChange() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(s.T(), resCtx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(testState.updateMask))
}

func (s *checkUpdateMaskSuite) TestCheckUpdateMaskNotModify() {
	fakeHttpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(s.T(), "unexpected request: "+r.URL.String())
	}))
	gcpNfsInstance := getGcpNfsInstanceWithoutStatus()

	factory, err := newTestStateFactory(fakeHttpServer, gcpNfsInstance)
	assert.Nil(s.T(), err)

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
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx := checkUpdateMask(ctx, testState.State)
	assert.Nil(s.T(), resCtx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(testState.updateMask))

	//ADD operation
	testState.operation = client.ADD
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx = checkUpdateMask(ctx, testState.State)
	assert.Nil(s.T(), resCtx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(testState.updateMask))

	//DELETE operation
	testState.operation = client.DELETE
	assert.Nil(s.T(), err)
	defer testState.FakeHttpServer.Close()
	err, resCtx = checkUpdateMask(ctx, testState.State)
	assert.Nil(s.T(), resCtx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(testState.updateMask))
}

func TestCheckUpdateMask(t *testing.T) {
	suite.Run(t, new(checkUpdateMaskSuite))
}
