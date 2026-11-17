package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func init() {
	// ArnPartition defaults to "" when the config package is not initialised.
	// Set it to the documented default so ARN assertions are deterministic.
	awsconfig.AwsConfig.ArnPartition = "aws"
}

// TestCloudWatchLogsStore_Lifecycle exercises the full happy-path of a log
// group: create → describe (prefix / pattern / identifier) → tag → list tags →
// add more tags → untag → delete.
func TestCloudWatchLogsStore_Lifecycle(t *testing.T) {
	const (
		account = "123456789012"
		region  = "eu-west-1"
		lgName  = "my-log-group"
	)

	ctx := context.Background()
	store := newCloudWatchLogsStore(account, region)

	// --- Create ---
	_, err := store.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: ptr.To(lgName),
		Tags:         map[string]string{"env": "test"},
	})
	require.NoError(t, err)

	// Derive the ARN the same way the store does so we can use it below.
	expectedArn := store.buildLogGroupArn(lgName)

	// --- Describe by prefix ---
	out, err := store.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: new("my-"),
	})
	require.NoError(t, err)
	require.Len(t, out.LogGroups, 1)
	assert.Equal(t, lgName, ptr.Deref(out.LogGroups[0].LogGroupName, ""))
	assert.Equal(t, expectedArn, ptr.Deref(out.LogGroups[0].LogGroupArn, ""))

	// --- Describe by pattern (substring) ---
	out, err = store.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePattern: new("log-group"),
	})
	require.NoError(t, err)
	require.Len(t, out.LogGroups, 1)

	// --- Describe by identifier (ARN without wildcard) ---
	out, err = store.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupIdentifiers: []string{expectedArn},
	})
	require.NoError(t, err)
	require.Len(t, out.LogGroups, 1)

	// --- Describe by identifier (ARN with wildcard suffix) ---
	out, err = store.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupIdentifiers: []string{expectedArn + ":*"},
	})
	require.NoError(t, err)
	require.Len(t, out.LogGroups, 1)

	// --- Describe with non-matching prefix returns nothing ---
	out, err = store.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: new("other-"),
	})
	require.NoError(t, err)
	assert.Empty(t, out.LogGroups)

	// --- List tags (initial tag set from creation) ---
	tagsOut, err := store.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
		ResourceArn: new(expectedArn),
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"env": "test"}, tagsOut.Tags)

	// --- TagResource: add more tags ---
	_, err = store.TagResource(ctx, &cloudwatchlogs.TagResourceInput{
		ResourceArn: new(expectedArn),
		Tags:        map[string]string{"owner": "cloud-manager", "env": "prod"},
	})
	require.NoError(t, err)

	tagsOut, err = store.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
		ResourceArn: new(expectedArn),
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"env": "prod", "owner": "cloud-manager"}, tagsOut.Tags)

	// --- UntagResource: remove one tag ---
	_, err = store.UntagResource(ctx, &cloudwatchlogs.UntagResourceInput{
		ResourceArn: new(expectedArn),
		TagKeys:     []string{"owner"},
	})
	require.NoError(t, err)

	tagsOut, err = store.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
		ResourceArn: new(expectedArn),
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"env": "prod"}, tagsOut.Tags)

	// --- Delete ---
	_, err = store.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: ptr.To(lgName),
	})
	require.NoError(t, err)

	// Describe after deletion must return empty result.
	out, err = store.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: ptr.To(lgName),
	})
	require.NoError(t, err)
	assert.Empty(t, out.LogGroups)
}

// TestCloudWatchLogsStore_SetLogGroupError verifies that every method that
// consults the error map returns the injected error, and that clearing the
// error (nil) restores normal behaviour.
func TestCloudWatchLogsStore_SetLogGroupError(t *testing.T) {
	const (
		account = "123456789012"
		region  = "eu-west-1"
		lgName  = "error-test-log-group"
	)

	ctx := context.Background()
	sentinel := errors.New("injected error")

	newStoreWithGroup := func() (*cloudWatchLogsStore, string) {
		s := newCloudWatchLogsStore(account, region)
		_, err := s.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
			LogGroupName: ptr.To(lgName),
		})
		require.NoError(t, err)
		return s, s.buildLogGroupArn(lgName)
	}

	t.Run("CreateLogGroup returns injected error", func(t *testing.T) {
		s := newCloudWatchLogsStore(account, region)
		lgArn := s.buildLogGroupArn(lgName)
		s.SetLogGroupError(lgArn, sentinel)

		_, err := s.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
			LogGroupName: ptr.To(lgName),
		})
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("DeleteLogGroup returns injected error", func(t *testing.T) {
		s, lgArn := newStoreWithGroup()
		s.SetLogGroupError(lgArn, sentinel)

		_, err := s.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: ptr.To(lgName),
		})
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("ListTagsForResource returns injected error", func(t *testing.T) {
		s, lgArn := newStoreWithGroup()
		s.SetLogGroupError(lgArn, sentinel)

		_, err := s.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
			ResourceArn: new(lgArn),
		})
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("TagResource returns injected error", func(t *testing.T) {
		s, lgArn := newStoreWithGroup()
		s.SetLogGroupError(lgArn, sentinel)

		_, err := s.TagResource(ctx, &cloudwatchlogs.TagResourceInput{
			ResourceArn: new(lgArn),
			Tags:        map[string]string{"k": "v"},
		})
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("UntagResource returns injected error", func(t *testing.T) {
		s, lgArn := newStoreWithGroup()
		s.SetLogGroupError(lgArn, sentinel)

		_, err := s.UntagResource(ctx, &cloudwatchlogs.UntagResourceInput{
			ResourceArn: new(lgArn),
			TagKeys:     []string{"k"},
		})
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("clearing error with nil restores normal behaviour", func(t *testing.T) {
		s, lgArn := newStoreWithGroup()
		s.SetLogGroupError(lgArn, sentinel)

		// Confirm it's broken first.
		_, err := s.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
			ResourceArn: new(lgArn),
		})
		require.ErrorIs(t, err, sentinel)

		// Clear it.
		s.SetLogGroupError(lgArn, nil)

		_, err = s.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
			ResourceArn: new(lgArn),
		})
		assert.NoError(t, err)
	})
}
