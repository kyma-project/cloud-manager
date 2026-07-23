package iprange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCidrOverlapsWith(t *testing.T) {
	const vpcCidr = "10.180.0.0/16"

	t.Run("no overlap when candidate is outside all forbidden ranges", func(t *testing.T) {
		overlap, err := cidrOverlapsWith("10.181.0.0/22", []string{vpcCidr})
		require.NoError(t, err)
		assert.Empty(t, overlap)
	})

	t.Run("overlap detected when candidate is a subnet of VPC primary CIDR", func(t *testing.T) {
		overlap, err := cidrOverlapsWith("10.180.64.0/22", []string{vpcCidr})
		require.NoError(t, err)
		assert.Equal(t, vpcCidr, overlap)
	})

	t.Run("overlap detected when candidate overlaps an existing secondary CIDR block", func(t *testing.T) {
		secondary := "10.181.0.0/20"
		overlap, err := cidrOverlapsWith("10.181.0.0/22", []string{vpcCidr, secondary})
		require.NoError(t, err)
		assert.Equal(t, secondary, overlap)
	})

	t.Run("no overlap when candidate is identical to an existing secondary block (idempotent re-association)", func(t *testing.T) {
		secondary := "10.181.0.0/22"
		overlap, err := cidrOverlapsWith("10.181.0.0/22", []string{vpcCidr, secondary})
		require.NoError(t, err)
		assert.Empty(t, overlap)
	})

	t.Run("error when candidate CIDR is unparseable", func(t *testing.T) {
		_, err := cidrOverlapsWith("not-a-cidr", []string{vpcCidr})
		require.Error(t, err)
	})

	t.Run("empty forbidden entries are skipped without panic", func(t *testing.T) {
		overlap, err := cidrOverlapsWith("10.181.0.0/22", []string{"", vpcCidr, ""})
		require.NoError(t, err)
		assert.Empty(t, overlap)
	})

	t.Run("unparseable forbidden entries are skipped without panic", func(t *testing.T) {
		overlap, err := cidrOverlapsWith("10.181.0.0/22", []string{"bad-cidr", vpcCidr})
		require.NoError(t, err)
		assert.Empty(t, overlap)
	})
}
