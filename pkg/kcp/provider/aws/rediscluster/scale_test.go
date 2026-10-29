package rediscluster

import (
	"context"
	"errors"
	"testing"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
)

func TestScaleElastiCacheClusterShards_NilRGSkips(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestState(t, "abc", false, nil, nil)
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterShards(context.Background(), state)

	assert.Nil(t, err, "must skip (nil) when replication group is not loaded")
	assert.Empty(t, stub.modifyShardConfigCalls, "must not call AWS when RG is nil")
}

func TestScaleElastiCacheClusterShards_UpToDateSkips(t *testing.T) {
	stub := newElastiCacheStub()
	rg := &elasticachetypes.ReplicationGroup{NodeGroups: []elasticachetypes.NodeGroup{{}}}
	state := newTestState(t, "abc", false, rg, nil)
	state.ObjAsRedisCluster().Spec.Instance.Aws.ShardCount = 1
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterShards(context.Background(), state)

	assert.Nil(t, err, "must skip (nil) when the shard count already matches")
	assert.Empty(t, stub.modifyShardConfigCalls, "must not call AWS when shard count is up to date")
}

func TestScaleElastiCacheClusterShards_ErrorSetsStateError(t *testing.T) {
	stub := newElastiCacheStub()
	stub.modifyShardConfigErr = errors.New("scale boom")
	rg := &elasticachetypes.ReplicationGroup{
		ReplicationGroupId: new("cm-abc"),
		NodeGroups:         []elasticachetypes.NodeGroup{{}},
	}
	state := newTestState(t, "abc", false, rg, nil)
	state.ObjAsRedisCluster().Spec.Instance.Aws.ShardCount = 2
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterShards(context.Background(), state)

	assert.NotNil(t, err, "AWS scale error must requeue, not continue")
	assert.NotEqual(t, composed.StopAndForget, err, "a scale error must requeue, not terminate")
	assert.Len(t, stub.modifyShardConfigCalls, 1, "AWS shard scaling must be attempted")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond, "Error condition must be set on scale failure")
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason)
}

func TestScaleElastiCacheClusterReplicas_NilRGSkips(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestState(t, "abc", false, nil, nil)
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterReplicas(context.Background(), state)

	assert.Nil(t, err, "must skip (nil) when replication group is not loaded")
	assert.Empty(t, stub.modifyReplicaConfigCalls, "must not call AWS when RG is nil")
}

func TestScaleElastiCacheClusterReplicas_UpToDateSkips(t *testing.T) {
	stub := newElastiCacheStub()
	rg := &elasticachetypes.ReplicationGroup{
		NodeGroups: []elasticachetypes.NodeGroup{{NodeGroupMembers: []elasticachetypes.NodeGroupMember{{}}}},
	}
	state := newTestState(t, "abc", false, rg, nil)
	state.ObjAsRedisCluster().Spec.Instance.Aws.ReplicasPerShard = 0
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterReplicas(context.Background(), state)

	assert.Nil(t, err, "must skip (nil) when the replica count already matches")
	assert.Empty(t, stub.modifyReplicaConfigCalls, "must not call AWS when replica count is up to date")
}

func TestScaleElastiCacheClusterReplicas_ErrorSetsStateError(t *testing.T) {
	stub := newElastiCacheStub()
	stub.modifyReplicaConfigErr = errors.New("scale boom")
	rg := &elasticachetypes.ReplicationGroup{
		ReplicationGroupId: new("cm-abc"),
		NodeGroups:         []elasticachetypes.NodeGroup{{}},
	}
	state := newTestState(t, "abc", false, rg, nil)
	state.ObjAsRedisCluster().Spec.Instance.Aws.ReplicasPerShard = 2
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterReplicas(context.Background(), state)

	assert.NotNil(t, err, "AWS scale error must requeue, not continue")
	assert.NotEqual(t, composed.StopAndForget, err, "a scale error must requeue, not terminate")
	assert.Len(t, stub.modifyReplicaConfigCalls, 1, "AWS replica scaling must be attempted")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond, "Error condition must be set on scale failure")
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason)
}

func TestScaleElastiCacheClusterShards_SuccessRequeues(t *testing.T) {
	stub := newElastiCacheStub() // no error
	rg := &elasticachetypes.ReplicationGroup{
		ReplicationGroupId: new("cm-abc"),
		NodeGroups:         []elasticachetypes.NodeGroup{{}},
	}
	state := newTestState(t, "abc", false, rg, nil)
	state.ObjAsRedisCluster().Spec.Instance.Aws.ShardCount = 2
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterShards(context.Background(), state)

	assert.Len(t, stub.modifyShardConfigCalls, 1, "AWS shard scaling must be attempted")
	assert.NotNil(t, err, "a successfully-initiated scale must requeue to await completion, not continue the pipeline")
	assert.NotEqual(t, composed.StopAndForget, err, "a successful scale must requeue, not terminate")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State, "a successful scale must not set StateError")
}

func TestScaleElastiCacheClusterReplicas_SuccessRequeues(t *testing.T) {
	stub := newElastiCacheStub() // no error
	rg := &elasticachetypes.ReplicationGroup{
		ReplicationGroupId: new("cm-abc"),
		NodeGroups:         []elasticachetypes.NodeGroup{{}},
	}
	state := newTestState(t, "abc", false, rg, nil)
	state.ObjAsRedisCluster().Spec.Instance.Aws.ReplicasPerShard = 2
	state.awsClient = stub

	err, _ := scaleElastiCacheClusterReplicas(context.Background(), state)

	assert.Len(t, stub.modifyReplicaConfigCalls, 1, "AWS replica scaling must be attempted")
	assert.NotNil(t, err, "a successfully-initiated scale must requeue to await completion, not continue the pipeline")
	assert.NotEqual(t, composed.StopAndForget, err, "a successful scale must requeue, not terminate")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State, "a successful scale must not set StateError")
}
