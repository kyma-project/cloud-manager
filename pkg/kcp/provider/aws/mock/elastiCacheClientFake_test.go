package mock

import (
	"context"
	"testing"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

// This suite locks the mock's semantics around auth-flip modifications so
// controller-test assertions have a reliable substrate. The real AWS client
// sets AuthTokenUpdateStrategy=Delete whenever UserGroupIdsToAdd is present,
// so the fake must flip AuthTokenEnabled to false in the same request. On
// the reverse (UserGroupIdsToRemove), auth stays as-is.
func TestElastiCacheFake_AuthTokenPropagation(t *testing.T) {
	newFakeWithAuthEnabledRG := func(name string) *elastiCacheClientFake {
		client := newElastiCacheClientFake()
		client.replicationGroups[name] = &elasticachetypes.ReplicationGroup{
			ReplicationGroupId: new(name),
			AuthTokenEnabled:   new(true),
			UserGroupIds:       []string{},
		}
		return client
	}

	t.Run("Add path flips AuthTokenEnabled to false and attaches UG (downgrade)", func(t *testing.T) {
		ctx := context.Background()
		client := newFakeWithAuthEnabledRG("cm-test")

		_, err := client.ModifyElastiCacheReplicationGroup(ctx, "cm-test", awsclient.ModifyElastiCacheClusterOptions{
			UserGroupIdsToAdd: []string{"cm-test"},
		})

		require.NoError(t, err)
		rg := client.GetAwsElastiCacheByName("cm-test")
		assert.False(t, ptr.Deref(rg.AuthTokenEnabled, true),
			"AuthTokenEnabled must flip to false when a user group is added (AuthTokenUpdateStrategy=Delete on the real client)")
		assert.Equal(t, []string{"cm-test"}, rg.UserGroupIds)
	})

	t.Run("Remove path detaches UG without altering AuthTokenEnabled", func(t *testing.T) {
		ctx := context.Background()
		// Seed AuthTokenEnabled=true to prove non-mutation: if Remove wrongly
		// touched the field, the assertion below would fail.
		client := newElastiCacheClientFake()
		client.replicationGroups["cm-test"] = &elasticachetypes.ReplicationGroup{
			ReplicationGroupId: new("cm-test"),
			AuthTokenEnabled:   new(true),
			UserGroupIds:       []string{"cm-test"},
		}

		_, err := client.ModifyElastiCacheReplicationGroup(ctx, "cm-test", awsclient.ModifyElastiCacheClusterOptions{
			UserGroupIdsToRemove: []string{"cm-test"},
		})

		require.NoError(t, err)
		rg := client.GetAwsElastiCacheByName("cm-test")
		assert.Empty(t, rg.UserGroupIds, "UG must be detached")
		assert.True(t, ptr.Deref(rg.AuthTokenEnabled, false),
			"AuthTokenEnabled must remain true — Remove must not mutate the auth flag")
	})

	t.Run("ModifyReplicationGroupCalls records each invocation in order", func(t *testing.T) {
		ctx := context.Background()
		client := newFakeWithAuthEnabledRG("cm-test")

		_, err := client.ModifyElastiCacheReplicationGroup(ctx, "cm-test", awsclient.ModifyElastiCacheClusterOptions{
			UserGroupIdsToAdd: []string{"cm-test"},
		})
		require.NoError(t, err)
		_, err = client.ModifyElastiCacheReplicationGroup(ctx, "cm-test", awsclient.ModifyElastiCacheClusterOptions{
			UserGroupIdsToRemove: []string{"cm-test"},
		})
		require.NoError(t, err)

		calls := client.ModifyReplicationGroupCalls()
		require.Len(t, calls, 2)
		assert.Equal(t, []string{"cm-test"}, calls[0].UserGroupIdsToAdd)
		assert.Equal(t, []string{"cm-test"}, calls[1].UserGroupIdsToRemove)
	})
}

func TestElastiCacheFake_DescribeReplicationGroupEvents(t *testing.T) {
	t.Run("unseeded id returns nil", func(t *testing.T) {
		client := newElastiCacheClientFake()
		events, err := client.DescribeElastiCacheReplicationGroupEvents(context.Background(), "no-such-id")
		require.NoError(t, err)
		assert.Nil(t, events)
	})

	t.Run("seeded events are returned", func(t *testing.T) {
		client := newElastiCacheClientFake()
		msg := "Insufficient memory"
		client.SetAwsElastiCacheEvents("cm-test", []elasticachetypes.Event{
			{Message: ptr.To(msg)},
		})

		events, err := client.DescribeElastiCacheReplicationGroupEvents(context.Background(), "cm-test")
		require.NoError(t, err)
		require.Len(t, events, 1)
		assert.Equal(t, msg, ptr.Deref(events[0].Message, ""))
	})

	t.Run("different ids are independent", func(t *testing.T) {
		client := newElastiCacheClientFake()
		client.SetAwsElastiCacheEvents("rg-a", []elasticachetypes.Event{{Message: ptr.To("a")}})

		events, err := client.DescribeElastiCacheReplicationGroupEvents(context.Background(), "rg-b")
		require.NoError(t, err)
		assert.Nil(t, events, "unseeded id rg-b must return nil, not rg-a's events")
	})
}
