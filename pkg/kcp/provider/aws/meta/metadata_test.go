package meta

import (
	"errors"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"testing"
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
