package redisinstance

import (
	"context"
	"errors"
	"testing"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

// elastiCacheStub is a minimal awsclient.ElastiCacheClient for action-guard
// unit tests. Only the methods the guards can plausibly call are implemented;
// every other method panics ("unimplemented") so a broken guard that lets a
// call slip through fails loudly instead of silently accepting a zero-return.
type elastiCacheStub struct {
	createUserGroupCalls   []string
	deleteUserGroupCalls   []string
	describeUserGroupCalls []string
	preloadedUserGroups    map[string]*elasticachetypes.UserGroup

	describeEventsReturns []elasticachetypes.Event
	describeEventsCalls   []string
	describeEventsErr     error
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
	panic("unimplemented")
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

func newTestStateWithClient(t *testing.T, instanceName string, ug *elasticachetypes.UserGroup, stub *elastiCacheStub) *State {
	t.Helper()
	state := newTestState(t, instanceName, false, &elasticachetypes.ReplicationGroup{}, ug)
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
	foreignName := "cm-different-instance"
	foreignUG := &elasticachetypes.UserGroup{
		UserGroupId: new(foreignName),
		Status:      ptr.To(string(awsmeta.ElastiCache_UserGroup_ACTIVE)),
	}
	stub.preloadedUserGroups[foreignName] = foreignUG

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
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisInstance().Status.State,
		"RedisInstance must not be set to StateError while member clusters are not yet visible")
	cond := meta.FindStatusCondition(state.ObjAsRedisInstance().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
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
			err, _ := waitElastiCacheAvailable(context.Background(), state)
			assert.NotNil(t, err, "waitElastiCacheAvailable must requeue while RG is %q", status)
			assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisInstance().Status.State,
				"non-terminal states must not set a terminal error state for %q", status)
		})
	}
}

func TestWaitElastiCacheAvailable_CreateFailedSurfacesError(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsReturns = []elasticachetypes.Event{
		{Message: ptr.To("InsufficientCacheClusterCapacity"), Date: ptr.To(time.Unix(3000, 0))},
	}
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err,
		"waitElastiCacheAvailable must StopAndForget when RG is create-failed")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisInstance().Status.State,
		"RedisInstance must be set to StateError when RG is create-failed")
	cond := meta.FindStatusCondition(state.ObjAsRedisInstance().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond, "Error condition must be set when RG is create-failed")
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason,
		"Error condition must use the CloudProviderError reason")
	assert.Contains(t, cond.Message, "InsufficientCacheClusterCapacity",
		"condition message must include the latest event reason")
	assert.NotContains(t, cond.Message, string(awsmeta.ElastiCache_CREATE_FAILED),
		"condition message must not expose the raw AWS replication group status")
	assert.Len(t, stub.describeEventsCalls, 1, "DescribeElastiCacheReplicationGroupEvents must be called once")
}

func TestWaitElastiCacheAvailable_CreateFailed_NoEvents(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsReturns = nil
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err,
		"must StopAndForget on create-failed with no events")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisInstance().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisInstance().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond)
	assert.Equal(t, "Failed to provision RedisInstance", cond.Message,
		"base message must be used when no events are available")
	assert.Len(t, stub.describeEventsCalls, 1, "DescribeElastiCacheReplicationGroupEvents must still be called")
}

func TestWaitElastiCacheAvailable_CreateFailed_EventsError(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsErr = errors.New("throttled: Rate exceeded")
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.awsClient = stub

	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err,
		"must StopAndForget on create-failed even when events API errors")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisInstance().Status.State)
	cond := meta.FindStatusCondition(state.ObjAsRedisInstance().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond)
	assert.NotContains(t, cond.Message, "throttled",
		"events API error must not appear in condition message")
	assert.NotContains(t, cond.Message, "Rate exceeded",
		"events API error must not appear in condition message")
	assert.NotContains(t, cond.Message, string(awsmeta.ElastiCache_CREATE_FAILED),
		"raw create-failed status must not appear in condition message")
	assert.Equal(t, "Failed to provision RedisInstance", cond.Message,
		"base message must be used when the events API errors")
}

func TestWaitElastiCacheAvailable_CreateFailed_PicksMostRecentEvent(t *testing.T) {
	stub := newElastiCacheStub()
	stub.describeEventsReturns = []elasticachetypes.Event{
		{Message: ptr.To("OlderEvent"), Date: ptr.To(time.Unix(1000, 0))},
		{Message: ptr.To("NewestEvent"), Date: ptr.To(time.Unix(3000, 0))},
		{Message: ptr.To("MiddleEvent"), Date: ptr.To(time.Unix(2000, 0))},
	}
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	state.awsClient = stub

	_, _ = waitElastiCacheAvailable(context.Background(), state)

	cond := meta.FindStatusCondition(state.ObjAsRedisInstance().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond)
	assert.Contains(t, cond.Message, "NewestEvent",
		"condition message must surface the most-recently dated event, not the last in the slice")
	assert.NotContains(t, cond.Message, "OlderEvent", "only the most-recent event reason must be surfaced")
	assert.NotContains(t, cond.Message, "MiddleEvent", "only the most-recent event reason must be surfaced")
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
	assert.NotEqual(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisInstance().Status.State,
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
