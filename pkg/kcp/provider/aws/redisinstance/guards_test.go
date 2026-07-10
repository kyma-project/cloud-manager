package redisinstance

import (
	"context"
	"testing"

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
}

func newElastiCacheStub() *elastiCacheStub {
	return &elastiCacheStub{preloadedUserGroups: map[string]*elasticachetypes.UserGroup{}}
}

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
	err, _ := waitElastiCacheAvailable(context.Background(), state)
	assert.Nil(t, err, "waitElastiCacheAvailable must proceed (nil) when RG is AVAILABLE")
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
	state := newTestState(t, "abc", false, &elasticachetypes.ReplicationGroup{
		Status: ptr.To(string(awsmeta.ElastiCache_CREATE_FAILED)),
	}, nil)
	err, _ := waitElastiCacheAvailable(context.Background(), state)

	assert.Equal(t, composed.StopAndForget, err,
		"waitElastiCacheAvailable must StopAndForget when RG is create-failed")
	assert.Equal(t, cloudcontrolv1beta1.StateError, state.ObjAsRedisInstance().Status.State,
		"RedisInstance must be set to StateError when RG is create-failed")
	cond := meta.FindStatusCondition(state.ObjAsRedisInstance().Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
	assert.NotNil(t, cond, "Error condition must be set when RG is create-failed")
	assert.Equal(t, cloudcontrolv1beta1.ReasonCloudProviderError, cond.Reason,
		"Error condition must use the CloudProviderError reason")
	assert.Equal(t, "Failed to provision RedisInstance", cond.Message,
		"condition message must stay abstract and not leak cloud provider internals")
	assert.NotContains(t, cond.Message, string(awsmeta.ElastiCache_CREATE_FAILED),
		"condition message must not expose the raw AWS replication group status")
}
