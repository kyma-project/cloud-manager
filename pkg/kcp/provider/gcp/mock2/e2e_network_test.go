package mock2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestE2ENetwork(t *testing.T) {

	t.Run("Network can not be deleted if used by PSA address range", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)
		net := s.createNetworkOK("test-net")
		_ = s.createPsaRangeOK(net.GetSelfLink(), "test-address", "10.251.0.0", 16)

		opVoid, err := s.deleteNetwork(net.GetName())
		require.Error(t, err)
		require.Contains(s.t, err.Error(), fmt.Sprintf("network projects/%[1]s/global/networks/test-net cannot be deleted because it has address projects/%[1]s/address/test-address", s.mock.ProjectId()))
		require.Nil(t, opVoid)
	})

}
