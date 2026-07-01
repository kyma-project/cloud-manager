package redisinstance

import (
	"context"
	"testing"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type testRedisInstanceState struct {
	focal.State
	ipRange *cloudcontrolv1beta1.IpRange
}

var _ types.State = &testRedisInstanceState{}

func (s *testRedisInstanceState) ObjAsRedisInstance() *cloudcontrolv1beta1.RedisInstance {
	return s.Obj().(*cloudcontrolv1beta1.RedisInstance)
}

func (s *testRedisInstanceState) IpRange() *cloudcontrolv1beta1.IpRange {
	return s.ipRange
}

func (s *testRedisInstanceState) SetIpRange(r *cloudcontrolv1beta1.IpRange) {
	s.ipRange = r
}

// newTestState builds a State backed by a real *RedisInstance with the given
// desiredAuth wired into spec.instance.aws.authEnabled. AWS-client wiring is
// intentionally omitted — callers under test never invoke a client method.
func newTestState(t *testing.T, instanceName string, desiredAuth bool, rg *elasticachetypes.ReplicationGroup, ug *elasticachetypes.UserGroup) *State {
	t.Helper()

	fakeClient := fake.NewClientBuilder().WithScheme(commonscheme.KcpScheme).Build()
	cluster := composed.NewStateCluster(fakeClient, fakeClient, nil, fakeClient.Scheme())

	obj := &cloudcontrolv1beta1.RedisInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: "kcp-system",
		},
		Spec: cloudcontrolv1beta1.RedisInstanceSpec{
			Instance: cloudcontrolv1beta1.RedisInstanceInfo{
				Aws: &cloudcontrolv1beta1.RedisInstanceAws{
					AuthEnabled: desiredAuth,
				},
			},
		},
	}

	focalState := focal.NewStateFactory().NewState(
		composed.NewStateFactory(cluster).NewState(k8stypes.NamespacedName{Name: instanceName, Namespace: "kcp-system"}, obj),
	)
	focalState.SetScope(&cloudcontrolv1beta1.Scope{})

	return &State{
		State:                       &testRedisInstanceState{State: focalState},
		elastiCacheReplicationGroup: rg,
		userGroup:                   ug,
	}
}

func TestShouldCreateTransientUserGroupPredicate(t *testing.T) {
	tests := []struct {
		name        string
		rgPresent   bool
		currentAuth bool
		desiredAuth bool
		ugPresent   bool
		expected    bool
	}{
		{"downgrade needed, UG missing", true, true, false, false, true},
		{"UG already present", true, true, false, true, false},
		{"no downgrade requested (upgrade)", true, true, true, false, false},
		{"already downgraded", true, false, false, false, false},
		{"RG not loaded yet", false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rg *elasticachetypes.ReplicationGroup
			if tt.rgPresent {
				rg = &elasticachetypes.ReplicationGroup{
					AuthTokenEnabled: ptr.To(tt.currentAuth),
				}
			}
			var ug *elasticachetypes.UserGroup
			if tt.ugPresent {
				ug = &elasticachetypes.UserGroup{
					UserGroupId: ptr.To("cm-test"),
					Status:      ptr.To("active"),
				}
			}
			state := newTestState(t, "test", tt.desiredAuth, rg, ug)

			got := shouldCreateTransientUserGroupPredicate()(context.Background(), state)
			assert.Equal(t, tt.expected, got, "predicate returned unexpected value for case %q", tt.name)
		})
	}
}

func TestShouldDeleteTransientUserGroupPredicate(t *testing.T) {
	// The predicate simplifies to `ugExists AND !ugAttached`.
	const instanceName = "test"
	attachedName := GetAwsElastiCacheUserGroupName(instanceName)

	tests := []struct {
		name         string
		ugPresent    bool
		attachedList []string
		currentAuth  bool
		desiredAuth  bool
		expected     bool
	}{
		{"orphan on auth-enabled — backfill delete", true, []string{}, true, true, true},
		{"post-detach: currentAuth=false, desiredAuth=false", true, []string{}, false, false, true},
		{"upgrade path with leftover UG", true, []string{}, false, true, true},
		{"mid-downgrade before Add propagated", true, []string{attachedName}, true, false, false},
		{"mid-downgrade after Add propagated, before Remove", true, []string{attachedName}, false, false, false},
		{"nothing to delete — ug nil", false, []string{}, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := &elasticachetypes.ReplicationGroup{
				AuthTokenEnabled: ptr.To(tt.currentAuth),
				UserGroupIds:     tt.attachedList,
			}
			var ug *elasticachetypes.UserGroup
			if tt.ugPresent {
				ug = &elasticachetypes.UserGroup{
					UserGroupId: ptr.To(attachedName),
					Status:      ptr.To("active"),
				}
			}
			state := newTestState(t, instanceName, tt.desiredAuth, rg, ug)

			got := shouldDeleteTransientUserGroupPredicate()(context.Background(), state)
			assert.Equal(t, tt.expected, got, "predicate returned unexpected value for case %q", tt.name)
		})
	}
}

func TestIsUserGroupAttached(t *testing.T) {
	const instanceName = "abc"
	attached := GetAwsElastiCacheUserGroupName(instanceName)

	tests := []struct {
		name         string
		attachedList []string
		expected     bool
	}{
		{"detached", []string{}, false},
		{"attached exactly", []string{attached}, true},
		{"other UG attached", []string{"cm-other"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newTestState(t, instanceName, false,
				&elasticachetypes.ReplicationGroup{UserGroupIds: tt.attachedList},
				&elasticachetypes.UserGroup{UserGroupId: ptr.To(attached)},
			)
			assert.Equal(t, tt.expected, isUserGroupAttached(state))
		})
	}
}
