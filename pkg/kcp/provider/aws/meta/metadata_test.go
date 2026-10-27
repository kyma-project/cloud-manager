package meta

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFoundResponseError(t *testing.T) {
	err := NewHttpNotFoundError(errors.New("not found error"))
	assert.True(t, IsNotFound(err))
}

func TestIsNotFound(t *testing.T) {
	codes := []string{
		"FileSystemNotFound",
		"AccessPointNotFound",
		"MountTargetNotFound",
		"PolicyNotFound",
		"CacheSubnetGroupNotFoundFault",
		"CacheClusterNotFound",
		"ResourceNotFoundException",
		"InvalidVpcID.NotFound",
		"InvalidVpcPeeringConnectionID.NotFound",
		"InvalidRoute.NotFound",
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			assert.True(t, IsNotFound(&smithy.GenericAPIError{Code: code}))
		})
	}
}

func TestEvaluateProvisioningAction(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	timeout := ElastiCacheProvisioningTimeout
	beforeDeadline := timeout - time.Second
	atDeadline := timeout
	afterDeadline := timeout + time.Second

	cases := []struct {
		name  string
		state string
		age   time.Duration
		want  ProvisioningAction
	}{
		// creating
		{"creating/before", ElastiCache_CREATING, beforeDeadline, ProvisioningActionWait},
		{"creating/at", ElastiCache_CREATING, atDeadline, ProvisioningActionTimeout},
		{"creating/after", ElastiCache_CREATING, afterDeadline, ProvisioningActionTimeout},
		// create-failed
		{"create-failed/before", ElastiCache_CREATE_FAILED, beforeDeadline, ProvisioningActionHeal},
		{"create-failed/at", ElastiCache_CREATE_FAILED, atDeadline, ProvisioningActionTimeout},
		{"create-failed/after", ElastiCache_CREATE_FAILED, afterDeadline, ProvisioningActionTimeout},
		// deleting — NEVER Timeout, even past the deadline
		{"deleting/before", ElastiCache_DELETING, beforeDeadline, ProvisioningActionWait},
		{"deleting/at", ElastiCache_DELETING, atDeadline, ProvisioningActionWait},
		{"deleting/after", ElastiCache_DELETING, afterDeadline, ProvisioningActionWait},
		// other states — the provisioning deadline must NEVER apply to an
		// already-provisioned instance (available/modifying/snapshotting),
		// regardless of age.
		{"available/before", ElastiCache_AVAILABLE, beforeDeadline, ProvisioningActionWait},
		{"available/at", ElastiCache_AVAILABLE, atDeadline, ProvisioningActionWait},
		{"available/after", ElastiCache_AVAILABLE, afterDeadline, ProvisioningActionWait},
		{"modifying/before", ElastiCache_MODIFYING, beforeDeadline, ProvisioningActionWait},
		{"modifying/at", ElastiCache_MODIFYING, atDeadline, ProvisioningActionWait},
		{"modifying/after", ElastiCache_MODIFYING, afterDeadline, ProvisioningActionWait},
		{"snapshotting/before", ElastiCache_SNAPSHOTTING, beforeDeadline, ProvisioningActionWait},
		{"snapshotting/at", ElastiCache_SNAPSHOTTING, atDeadline, ProvisioningActionWait},
		{"snapshotting/after", ElastiCache_SNAPSHOTTING, afterDeadline, ProvisioningActionWait},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			createdAt := now.Add(-tc.age)
			got := EvaluateProvisioningAction(tc.state, createdAt, now)
			assert.Equal(t, tc.want, got, "state=%s age=%s", tc.state, tc.age)
		})
	}
}
