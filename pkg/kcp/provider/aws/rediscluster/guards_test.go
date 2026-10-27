package rediscluster

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// elastiCacheStub is a minimal awsclient.ElastiCacheClient implementation for
// action-guard unit tests. Only the methods the guards can plausibly call are
// implemented; every other method panics ("unimplemented") so that if a
// guard's early-return path is broken and a call slips through, the test
// fails loudly instead of silently accepting a zero-return.
type elastiCacheStub struct {
	createUserGroupCalls   []string
	deleteUserGroupCalls   []string
	describeUserGroupCalls []string
	// preloadedUserGroups: DescribeUserGroup will return these unchanged.
	// Used to prove foreign UG names remain untouched.
	preloadedUserGroups map[string]*elasticachetypes.UserGroup

	describeEventsReturns []elasticachetypes.Event
	describeEventsCalls   []string
	describeEventsErr     error

	deleteReplicationGroupCalls []string
	deleteReplicationGroupErr   error
}

func newElastiCacheStub() *elastiCacheStub {
	return &elastiCacheStub{preloadedUserGroups: map[string]*elasticachetypes.UserGroup{}}
}

// double-waiver: B1 — ElastiCacheClient wraps AWS ElastiCache API calls (out-of-process network handle)
var _ awsclient.ElastiCacheClient = &elastiCacheStub{}

func (s *elastiCacheStub) CreateUserGroup(ctx context.Context, id string, tags []elasticachetypes.Tag) (*elasticache.CreateUserGroupOutput, error) {
	s.createUserGroupCalls = append(s.createUserGroupCalls, id)
	return &elasticache.CreateUserGroupOutput{UserGroupId: new(id)}, nil
}

func (s *elastiCacheStub) DeleteUserGroup(ctx context.Context, id string) error {
	s.deleteUserGroupCalls = append(s.deleteUserGroupCalls, id)
	return nil
}

func (s *elastiCacheStub) DescribeUserGroup(ctx context.Context, id string) (*elasticachetypes.UserGroup, error) {
	s.describeUserGroupCalls = append(s.describeUserGroupCalls, id)
	return s.preloadedUserGroups[id], nil
}

// All remaining methods must panic — the guards under test should never
// reach them. A silent nil return would mask a broken guard.
func (s *elastiCacheStub) DescribeElastiCacheSubnetGroup(ctx context.Context, name string) ([]elasticachetypes.CacheSubnetGroup, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) CreateElastiCacheSubnetGroup(ctx context.Context, name string, subnetIds []string, tags []elasticachetypes.Tag) (*elasticache.CreateCacheSubnetGroupOutput, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) DeleteElastiCacheSubnetGroup(ctx context.Context, name string) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) DescribeElastiCacheParameterGroup(ctx context.Context, name string) ([]elasticachetypes.CacheParameterGroup, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) CreateElastiCacheParameterGroup(ctx context.Context, name, family string, tags []elasticachetypes.Tag) (*elasticache.CreateCacheParameterGroupOutput, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) DeleteElastiCacheParameterGroup(ctx context.Context, name string) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) DescribeElastiCacheParameters(ctx context.Context, groupName string) ([]elasticachetypes.Parameter, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) ModifyElastiCacheParameterGroup(ctx context.Context, groupName string, parameters []elasticachetypes.ParameterNameValue) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) DescribeEngineDefaultParameters(ctx context.Context, family string) ([]elasticachetypes.Parameter, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) GetAuthTokenSecretValue(ctx context.Context, secretName string) (*secretsmanager.GetSecretValueOutput, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) CreateAuthTokenSecret(ctx context.Context, secretName string, tags []secretsmanagertypes.Tag) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) DeleteAuthTokenSecret(ctx context.Context, secretName string) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) DescribeElastiCacheReplicationGroup(ctx context.Context, clusterId string) ([]elasticachetypes.ReplicationGroup, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) CreateElastiCacheReplicationGroup(ctx context.Context, tags []elasticachetypes.Tag, options awsclient.CreateElastiCacheClusterOptions) (*elasticache.CreateReplicationGroupOutput, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) ModifyElastiCacheReplicationGroup(ctx context.Context, id string, options awsclient.ModifyElastiCacheClusterOptions) (*elasticache.ModifyReplicationGroupOutput, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) DeleteElastiCacheReplicationGroup(ctx context.Context, id string) error {
	s.deleteReplicationGroupCalls = append(s.deleteReplicationGroupCalls, id)
	return s.deleteReplicationGroupErr
}
func (s *elastiCacheStub) DescribeElastiCacheCluster(ctx context.Context, id string) ([]elasticachetypes.CacheCluster, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) ModifyElastiCacheClusterShardConfiguration(ctx context.Context, options awsclient.RescaleElastiCacheClusterShardOptions) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) ModifyElastiCacheClusterReplicaConfiguration(ctx context.Context, options awsclient.RescaleElastiCacheClusterReplicaOptions) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) DescribeElastiCacheSecurityGroups(ctx context.Context, filters []ec2types.Filter, groupIds []string) ([]ec2types.SecurityGroup, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) CreateElastiCacheSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2types.Tag) (string, error) {
	panic("unimplemented")
}
func (s *elastiCacheStub) DeleteElastiCacheSecurityGroup(ctx context.Context, id string) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) AuthorizeElastiCacheSecurityGroupIngress(ctx context.Context, groupId string, ipPermissions []ec2types.IpPermission) error {
	panic("unimplemented")
}
func (s *elastiCacheStub) DescribeElastiCacheReplicationGroupEvents(ctx context.Context, replicationGroupId string) ([]elasticachetypes.Event, error) {
	s.describeEventsCalls = append(s.describeEventsCalls, replicationGroupId)
	return s.describeEventsReturns, s.describeEventsErr
}

// newTestStateWithClient wraps newTestState by attaching a stub AWS client so
// action guards under test can be exercised without hitting the panicking
// path. The stub records calls so tests can assert the guard actually
// short-circuited without issuing AWS calls.
func newTestStateWithClient(t *testing.T, clusterName string, ug *elasticachetypes.UserGroup, stub *elastiCacheStub) *State {
	t.Helper()
	state := newTestState(t, clusterName, false, &elasticachetypes.ReplicationGroup{}, ug)
	state.awsClient = stub
	return state
}

func TestCreateUserGroupGuard_UGAlreadyPresent(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestStateWithClient(t, "abc", &elasticachetypes.UserGroup{
		UserGroupId: new("cm-abc"),
		Status:      ptr.To(string(awsmeta.ElastiCache_UserGroup_ACTIVE)),
	}, stub)

	err, _ := createUserGroup(context.Background(), state)

	assert.Nil(t, err, "createUserGroup must return nil when userGroup is already loaded")
	assert.Empty(t, stub.createUserGroupCalls, "createUserGroup must not invoke AWS when userGroup is present")
}

func TestDeleteUserGroupGuard_UGNil(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestStateWithClient(t, "abc", nil, stub)

	err, _ := deleteUserGroup(context.Background(), state)

	assert.Nil(t, err)
	assert.Empty(t, stub.deleteUserGroupCalls, "deleteUserGroup must not invoke AWS when userGroup is nil")
}

func TestDeleteUserGroupGuard_UGDeleting(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestStateWithClient(t, "abc", &elasticachetypes.UserGroup{
		UserGroupId: new("cm-abc"),
		Status:      ptr.To(string(awsmeta.ElastiCache_UserGroup_DELETING)),
	}, stub)

	err, _ := deleteUserGroup(context.Background(), state)

	assert.Nil(t, err)
	assert.Empty(t, stub.deleteUserGroupCalls, "deleteUserGroup must skip AWS when UG is already DELETING")
}

// Defensive guard against a fresh-create → downgrade race: if the transient
// user group was just created and hasn't reached ACTIVE yet, AWS rejects
// DeleteUserGroup on a CREATING resource. The guard must instead requeue.
func TestDeleteUserGroupGuard_UGCreating(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestStateWithClient(t, "abc", &elasticachetypes.UserGroup{
		UserGroupId: new("cm-abc"),
		Status:      ptr.To(string(awsmeta.ElastiCache_UserGroup_CREATING)),
	}, stub)

	err, _ := deleteUserGroup(context.Background(), state)

	assert.NotNil(t, err, "deleteUserGroup must requeue when UG is still CREATING; got nil")
	assert.Empty(t, stub.deleteUserGroupCalls, "deleteUserGroup must not invoke AWS on a CREATING UG")
}

func TestWaitUserGroupActive_UGNil(t *testing.T) {
	state := newTestStateWithClient(t, "abc", nil, newElastiCacheStub())
	err, _ := waitUserGroupActive(context.Background(), state)
	assert.Nil(t, err, "waitUserGroupActive must return nil when userGroup is nil (continues pipeline)")
}

func TestWaitUserGroupActive_UGActive(t *testing.T) {
	state := newTestStateWithClient(t, "abc", &elasticachetypes.UserGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_UserGroup_ACTIVE)),
	}, newElastiCacheStub())
	err, _ := waitUserGroupActive(context.Background(), state)
	assert.Nil(t, err, "waitUserGroupActive must return nil when UG is ACTIVE (continues pipeline)")
}

func TestWaitUserGroupActive_UGCreating(t *testing.T) {
	state := newTestStateWithClient(t, "abc", &elasticachetypes.UserGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_UserGroup_CREATING)),
	}, newElastiCacheStub())
	err, _ := waitUserGroupActive(context.Background(), state)
	assert.NotNil(t, err, "waitUserGroupActive must requeue when UG is CREATING")
}

// Foreign UG safety by construction. loadUserGroup fetches by exact
// GetAwsElastiCacheUserGroupName(state.Obj().GetName()); a differently-named
// user group in the same account must not be queried or touched.
func TestLoadUserGroup_ForeignNameUntouched(t *testing.T) {
	stub := newElastiCacheStub()
	// Pre-seed a foreign UG that must not be affected.
	foreignName := "cm-different-cluster"
	foreignUG := &elasticachetypes.UserGroup{
		UserGroupId: new(foreignName),
		Status:      ptr.To(string(awsmeta.ElastiCache_UserGroup_ACTIVE)),
	}
	stub.preloadedUserGroups[foreignName] = foreignUG

	// Also pre-seed our expected name so the load succeeds.
	ourName := GetAwsElastiCacheUserGroupName("abc")
	stub.preloadedUserGroups[ourName] = &elasticachetypes.UserGroup{
		UserGroupId: new(ourName),
		Status:      ptr.To(string(awsmeta.ElastiCache_UserGroup_ACTIVE)),
	}

	state := newTestState(t, "abc", true, &elasticachetypes.ReplicationGroup{}, nil)
	state.awsClient = stub

	err, _ := loadUserGroup(context.Background(), state)

	assert.Nil(t, err)
	assert.Equal(t, []string{ourName}, stub.describeUserGroupCalls,
		"loadUserGroup must query exactly our derived name — never a foreign name")
	// Foreign UG is unchanged in the stub's inventory.
	assert.Same(t, foreignUG, stub.preloadedUserGroups[foreignName],
		"foreign user group must remain untouched")
	assert.Equal(t, awsmeta.ElastiCache_UserGroup_ACTIVE, awsmeta.ElastiCacheUserGroupState(ptr.Deref(foreignUG.Status, "")))
}

func TestCreateUserGroupGuard_ProceedsWhenUGNil(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestStateWithClient(t, "abc", nil, stub)

	_, _ = createUserGroup(context.Background(), state)

	assert.Equal(t, []string{GetAwsElastiCacheUserGroupName("abc")}, stub.createUserGroupCalls,
		"createUserGroup must invoke AWS with the derived cm-<name> id when userGroup is nil")
}

func TestDeleteUserGroupGuard_ProceedsWhenActive(t *testing.T) {
	stub := newElastiCacheStub()
	id := GetAwsElastiCacheUserGroupName("abc")
	state := newTestStateWithClient(t, "abc", &elasticachetypes.UserGroup{
		UserGroupId: new(id),
		Status:      ptr.To(string(awsmeta.ElastiCache_UserGroup_ACTIVE)),
	}, stub)

	_, _ = deleteUserGroup(context.Background(), state)

	assert.Equal(t, []string{id}, stub.deleteUserGroupCalls,
		"deleteUserGroup must invoke AWS with the UG id when the UG is ACTIVE")
}

func TestWaitElastiCacheAvailable_Available(t *testing.T) {
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_AVAILABLE)),
	}, nil)
	state.memberClusters = []elasticachetypes.CacheCluster{{}}
	err, _ := waitElastiCacheAvailable(context.Background(), state)
	assert.Nil(t, err, "waitElastiCacheAvailable must proceed (nil) when RG is AVAILABLE with member clusters")
}

// available but zero member clusters is a normal eventual-consistency window
// during provisioning - requeue and wait for the members to appear rather than
// terminating (an AVAILABLE replication group's members become visible shortly
// after its status flips).
func TestWaitElastiCacheAvailable_AvailableButNoMembersRequeues(t *testing.T) {
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_AVAILABLE)),
	}, nil)
	state.memberClusters = nil
	err, _ := waitElastiCacheAvailable(context.Background(), state)
	assert.NotNil(t, err,
		"waitElastiCacheAvailable must requeue (non-nil) when AVAILABLE but no member clusters yet")
	assert.NotEqual(t, composed.StopAndForget, err,
		"waitElastiCacheAvailable must not terminate; member clusters should appear on a later reconcile")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State,
		"RedisCluster must not be set to StateError while member clusters are not yet visible")
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.Nil(t, cond, "Error condition must not be set for a transient no-members state")
}

func TestWaitElastiCacheAvailable_NonTerminalStatesRequeue(t *testing.T) {
	for _, status := range []string{
		awsmeta.ElastiCache_CREATING,
		awsmeta.ElastiCache_MODIFYING,
		awsmeta.ElastiCache_SNAPSHOTTING,
		awsmeta.ElastiCache_DELETING,
		"some-unexpected-status",
		"",
	} {
		t.Run(status, func(t *testing.T) {
			state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
				Status: new(status),
			}, nil)
			state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now()))
			err, _ := waitElastiCacheAvailable(context.Background(), state)
			assert.NotNil(t, err, "waitElastiCacheAvailable must requeue while RG is %q", status)
			assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State,
				"non-terminal states must not set a terminal error state for %q", status)
		})
	}
}

func TestWaitElastiCacheAvailable_CreateFailedYoung_Heals(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsReturns = []elasticachetypes.Event{
		{Message: ptr.To("InsufficientCacheClusterCapacity"), Date: ptr.To(time.Unix(3000, 0))},
	}
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now()))
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.NotNil(t, err, "create-failed within deadline must requeue, not terminate")
	assert.NotEqual(t, composed.StopAndForget, err, "must not StopAndForget within deadline")
	assert.Len(t, stub.deleteReplicationGroupCalls, 1, "delete must be called exactly once")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State,
		"must not set StateError during auto-heal")
	updatingCond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeUpdating)
	assert.NotNil(t, updatingCond, "ConditionTypeUpdating must be set during heal")
	errCond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.Nil(t, errCond, "ConditionTypeError must not be set during heal")
	assert.Len(t, stub.describeEventsCalls, 1, "DescribeElastiCacheReplicationGroupEvents must be called once")
}

func TestWaitElastiCacheAvailable_CreateFailedYoung_EventsError_HealProceeds(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsErr = errors.New("throttled: Rate exceeded")
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now()))
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.NotNil(t, err, "heal must requeue even when events API errors")
	assert.NotEqual(t, composed.StopAndForget, err, "must not terminate on events API error")
	assert.Len(t, stub.deleteReplicationGroupCalls, 1, "delete must be called even when events API errors")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State,
		"must not set StateError during heal")
	errCond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.Nil(t, errCond, "ConditionTypeError must not be set during heal")
}

func TestWaitElastiCacheAvailable_CreateFailedYoung_IdempotentHeal(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now()))
	state.awsClient = stub

	// First observation: heal fires
	err1, _ := waitElastiCacheAvailable(context.Background(), state)
	assert.NotNil(t, err1, "first observation must requeue")
	assert.NotEqual(t, composed.StopAndForget, err1)
	assert.Len(t, stub.deleteReplicationGroupCalls, 1)

	// The fake K8s client zeroes creationTimestamp during Status().Update(); reset it so the
	// second call still sees the object as young and routes to Heal, not Timeout.
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now()))

	// Second observation (RG still create-failed before AWS completes delete):
	// heal is idempotent — delete issued again, still requeues, never terminal
	err2, _ := waitElastiCacheAvailable(context.Background(), state)
	assert.NotNil(t, err2, "second observation must also requeue")
	assert.NotEqual(t, composed.StopAndForget, err2)
	assert.Len(t, stub.deleteReplicationGroupCalls, 2, "delete must be called on each heal observation")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
}

func TestWaitElastiCacheAvailable_CreateFailedYoung_DeleteError_Requeues(t *testing.T) {
	stub := newElastiCacheStub()
	stub.deleteReplicationGroupErr = errors.New("delete boom")
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now()))
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.NotNil(t, err, "delete error must requeue, not terminate")
	assert.NotEqual(t, composed.StopAndForget, err, "must not terminate on delete error")
	assert.Len(t, stub.deleteReplicationGroupCalls, 1, "delete must be attempted")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State,
		"must not set StateError on delete error")
	errCond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.Nil(t, errCond, "ConditionTypeError must not be set on delete error")
}

func TestWaitElastiCacheAvailable_CreateFailedOldTimestamp_TimeoutPicksMostRecentEvent(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsReturns = []elasticachetypes.Event{
		{Message: ptr.To("OlderEvent"), Date: ptr.To(time.Unix(1000, 0))},
		{Message: ptr.To("NewestEvent"), Date: ptr.To(time.Unix(3000, 0))},
		{Message: ptr.To("MiddleEvent"), Date: ptr.To(time.Unix(2000, 0))},
	}
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now().Add(-3 * time.Hour)))
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err, "create-failed past deadline must StopAndForget")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond)
	assert.Contains(t, cond.Message, "NewestEvent",
		"timeout message must surface the most-recently dated event, not the last in the slice")
	assert.NotContains(t, cond.Message, "OlderEvent", "only the most-recent event reason must be surfaced")
	assert.NotContains(t, cond.Message, "MiddleEvent", "only the most-recent event reason must be surfaced")
}

// Regression: an already-provisioned cluster that is modifying/snapshotting
// past the 2h mark must NOT be failed by the provisioning deadline — the
// deadline applies only to in-progress provisioning (creating/create-failed).
func TestWaitElastiCacheAvailable_ModifyingSnapshottingPastDeadline_Requeues(t *testing.T) {
	for _, status := range []string{awsmeta.ElastiCache_MODIFYING, awsmeta.ElastiCache_SNAPSHOTTING} {
		t.Run(status, func(t *testing.T) {
			state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
				Status: ptr.To(status),
			}, nil)
			state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now().Add(-3 * time.Hour)))
			err, _ := waitElastiCacheAvailable(context.Background(), state)
			assert.NotNil(t, err, "%s past the deadline must requeue", status)
			assert.NotEqual(t, composed.StopAndForget, err, "%s must not terminate past the deadline", status)
			assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State,
				"an already-provisioned %s cluster must not be failed by the provisioning deadline", status)
			cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
			assert.Nil(t, cond, "%s past deadline must not set an Error condition", status)
		})
	}
}

// loadMemberClusters must continue (not error) on an empty member list so
// waitElastiCacheAvailable can decide readiness/terminal state.
func TestLoadMemberClusters_NoMembersDefersToWait(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status:         ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
		MemberClusters: nil,
	}, nil)
	state.awsClient = stub

	err, _ := loadMemberClusters(context.Background(), state)

	assert.Nil(t, err, "loadMemberClusters must continue (nil) on an empty member list; waitElastiCacheAvailable owns state handling")
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State,
		"loadMemberClusters must not set a terminal error state")
}

// modifyPreferredMaintenanceWindow must skip (nil), not error, when there are
// no member clusters to modify.
func TestModifyPreferredMaintenanceWindow_NoMembersSkips(t *testing.T) {
	stub := newElastiCacheStub()
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status:         ptr.To(string(awsmeta.ElastiCache_AVAILABLE)),
		MemberClusters: nil,
	}, nil)
	state.awsClient = stub

	err, _ := modifyPreferredMaintenanceWindow(context.Background(), state)

	assert.Nil(t, err, "modifyPreferredMaintenanceWindow must skip (nil) when there are no member clusters")
}

// provisioning deadline tests — use ages well outside ±1min to avoid flakiness
// on slow machines; exact-boundary logic is covered by metadata_test.go.

func TestWaitElastiCacheAvailable_CreatingOldTimestamp_Timeout(t *testing.T) {
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(awsmeta.ElastiCache_CREATING),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now().Add(-3 * time.Hour)))
	state.awsClient = newElastiCacheStub()

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err, "timeout must StopAndForget")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond)
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason)
	assert.Contains(t, cond.Message, "delete and recreate this resource to retry")
	assert.Contains(t, cond.Message, "2h0m0s")
	assert.NotContains(t, cond.Message, string(awsmeta.ElastiCache_CREATE_FAILED))
}

func TestWaitElastiCacheAvailable_CreatingRecentTimestamp_Requeues(t *testing.T) {
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(awsmeta.ElastiCache_CREATING),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now()))

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.NotNil(t, err, "within-deadline creating must requeue")
	assert.NotEqual(t, composed.StopAndForget, err)
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
}

func TestWaitElastiCacheAvailable_DeletingOldTimestamp_Requeues(t *testing.T) {
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(awsmeta.ElastiCache_DELETING),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now().Add(-3 * time.Hour)))

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.NotNil(t, err, "in-progress delete must never be timed out")
	assert.NotEqual(t, composed.StopAndForget, err)
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
}

func TestWaitElastiCacheAvailable_CreatingAtDeadline_Timeout(t *testing.T) {
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(awsmeta.ElastiCache_CREATING),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now().Add(-awsmeta.ElastiCacheProvisioningTimeout)))
	state.awsClient = newElastiCacheStub()

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err, "age==deadline must be terminal")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond)
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason)
	assert.Contains(t, cond.Message, "delete and recreate this resource to retry",
		"boundary timeout message must carry the remediation")
}

func TestWaitElastiCacheAvailable_CreateFailedOldTimestamp_TimeoutWithReason(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsReturns = []elasticachetypes.Event{
		{Message: ptr.To("InsufficientCacheClusterCapacity"), Date: ptr.To(time.Unix(3000, 0))},
	}
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.Obj().SetCreationTimestamp(metav1.NewTime(time.Now().Add(-3 * time.Hour)))
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err, "create-failed past deadline must StopAndForget")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisCluster().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisCluster().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond)
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason)
	assert.Contains(t, cond.Message, "InsufficientCacheClusterCapacity")
	assert.Contains(t, cond.Message, "delete and recreate this resource to retry")
	assert.NotContains(t, cond.Message, string(awsmeta.ElastiCache_CREATE_FAILED))
	assert.Len(t, stub.describeEventsCalls, 1, "events must be fetched for create-failed timeout")
}
