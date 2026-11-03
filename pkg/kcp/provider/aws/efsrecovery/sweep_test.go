package efsrecovery

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/go-logr/logr"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

// testLabels satisfies the shoot/region/mode label set used by sweep metrics.
var testLabels = prometheus.Labels{"shoot": "test-shoot", "region": "us-east-1", "mode": "on"}

// baseClient is an EfsClient stub where all methods panic with "not implemented".
// Test-specific clients embed it and override only the methods they exercise.
type baseClient struct{}

func (baseClient) DescribeFileSystems(context.Context) ([]efstypes.FileSystemDescription, error) {
	panic("not implemented")
}
func (baseClient) CreateFileSystem(context.Context, string, efstypes.PerformanceMode, efstypes.ThroughputMode, []efstypes.Tag) (*efs.CreateFileSystemOutput, error) {
	panic("not implemented")
}
func (baseClient) DeleteFileSystem(context.Context, string) error { panic("not implemented") }
func (baseClient) DescribeMountTargets(context.Context, string) ([]efstypes.MountTargetDescription, error) {
	panic("not implemented")
}
func (baseClient) CreateMountTarget(context.Context, string, string, []string) (string, error) {
	panic("not implemented")
}
func (baseClient) DeleteMountTarget(context.Context, string) error { panic("not implemented") }
func (baseClient) DescribeMountTargetSecurityGroups(context.Context, string) ([]string, error) {
	panic("not implemented")
}

var _ awsclient.EfsClient = baseClient{}

// emptyRefresher returns an empty keeper set with no error.
func emptyRefresher(_ context.Context) (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

// errorRefresher returns an error.
func errorRefresher(_ context.Context) (map[string]struct{}, error) {
	return nil, errors.New("kcp list error")
}

func keeperRefresher(ids ...string) statusIdRefresher {
	return func(_ context.Context) (map[string]struct{}, error) {
		m := map[string]struct{}{}
		for _, id := range ids {
			m[id] = struct{}{}
		}
		return m, nil
	}
}

// --- deleteFileSystems tests ---

func TestDeleteFileSystems_CleanDelete(t *testing.T) {
	orphan := mkEfs("fs-orphan", nil, 0)

	remaining := deleteFileSystems(
		context.Background(),
		logr.Discard(),
		&deleteOnlyClient{deleteErr: nil},
		[]efstypes.FileSystemDescription{orphan},
		emptyRefresher,
		0,
		testLabels,
	)

	assert.Empty(t, remaining, "clean delete: orphan should not be in remaining")
}

// deleteOnlyClient overrides only DeleteFileSystem; all other methods panic via baseClient.
type deleteOnlyClient struct {
	baseClient
	deleteErr    error
	deleteCalled *bool
}

func (c *deleteOnlyClient) DeleteFileSystem(_ context.Context, _ string) error {
	if c.deleteCalled != nil {
		*c.deleteCalled = true
	}
	return c.deleteErr
}

// describeMtClient overrides DescribeMountTargets and DeleteMountTarget.
type describeMtClient struct {
	baseClient
	mts       []efstypes.MountTargetDescription
	err       error
	mtDeleted []string
}

func (c *describeMtClient) DescribeMountTargets(_ context.Context, _ string) ([]efstypes.MountTargetDescription, error) {
	return c.mts, c.err
}
func (c *describeMtClient) DeleteMountTarget(_ context.Context, mtId string) error {
	c.mtDeleted = append(c.mtDeleted, mtId)
	return nil
}

func TestDeleteFileSystems_BecameKeeper(t *testing.T) {
	const fsId = "fs-keeper"
	called := false
	orphan := mkEfs(fsId, nil, 0)

	remaining := deleteFileSystems(
		context.Background(),
		logr.Discard(),
		&deleteOnlyClient{deleteCalled: &called},
		[]efstypes.FileSystemDescription{orphan},
		keeperRefresher(fsId),
		0,
		testLabels,
	)

	assert.Empty(t, remaining, "becameKeeper: not in remaining (was skipped)")
	assert.False(t, called, "becameKeeper: DeleteFileSystem must not be called")
}

func TestDeleteFileSystems_FileSystemInUse(t *testing.T) {
	orphan := mkEfs("fs-inuse", nil, 0)

	remaining := deleteFileSystems(
		context.Background(),
		logr.Discard(),
		&deleteOnlyClient{deleteErr: &efstypes.FileSystemInUse{}},
		[]efstypes.FileSystemDescription{orphan},
		emptyRefresher,
		0,
		testLabels,
	)

	assert.Len(t, remaining, 1, "FileSystemInUse: orphan must stay in remaining for retry")
	assert.Equal(t, ptr.Deref(orphan.FileSystemId, ""), ptr.Deref(remaining[0].FileSystemId, ""))
}

func TestDeleteFileSystems_Throttling(t *testing.T) {
	orphan := mkEfs("fs-throttled", nil, 0)

	remaining := deleteFileSystems(
		context.Background(),
		logr.Discard(),
		&deleteOnlyClient{deleteErr: &efstypes.ThrottlingException{}},
		[]efstypes.FileSystemDescription{orphan},
		emptyRefresher,
		0,
		testLabels,
	)

	assert.Len(t, remaining, 1, "throttling: orphan must stay in remaining for retry")
}

func TestDeleteFileSystems_RefresherError(t *testing.T) {
	called := false
	orphan := mkEfs("fs-any", nil, 0)

	remaining := deleteFileSystems(
		context.Background(),
		logr.Discard(),
		&deleteOnlyClient{deleteCalled: &called},
		[]efstypes.FileSystemDescription{orphan},
		errorRefresher,
		0,
		testLabels,
	)

	assert.Len(t, remaining, 1, "refresher error: orphan skipped to remaining")
	assert.False(t, called, "refresher error: DeleteFileSystem must not be called")
}

// --- deleteMountTargets tests ---

func TestDeleteMountTargets_HappyPath(t *testing.T) {
	orphan1 := mkEfs("fs-1", nil, 1)
	orphan1.FileSystemId = new("fs-1")
	orphan2 := mkEfs("fs-2", nil, 1)
	orphan2.FileSystemId = new("fs-2")

	client := &describeMtClient{
		mts: []efstypes.MountTargetDescription{
			{MountTargetId: new("mt-a"), FileSystemId: new("fs-any")},
		},
	}

	deleteMountTargets(
		context.Background(),
		logr.Discard(),
		client,
		[]efstypes.FileSystemDescription{orphan1, orphan2},
		0,
		testLabels,
	)

	assert.Len(t, client.mtDeleted, 2, "happy path: one DeleteMountTarget per orphan")
}

func TestDeleteMountTargets_DescribeError(t *testing.T) {
	orphan := mkEfs("fs-err", nil, 0)
	client := &describeMtClient{err: errors.New("aws error")}

	// Must not panic; error is logged and orphan is skipped.
	assert.NotPanics(t, func() {
		deleteMountTargets(
			context.Background(),
			logr.Discard(),
			client,
			[]efstypes.FileSystemDescription{orphan},
			0,
			testLabels,
		)
	})
	assert.Empty(t, client.mtDeleted)
}
