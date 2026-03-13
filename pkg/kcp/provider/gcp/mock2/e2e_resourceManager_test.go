package mock2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/require"
)

func TestE2EResourceManager(t *testing.T) {

	t.Run("Resource manager tag keys, values and bindings used on network", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		s := newE2ETestSuite(ctx, t)
		net := s.createNetworkOK("test-net")

		netName, err := gcputil.ParseNameDetail(net.GetSelfLink())
		netName.PrefixWithGoogleApisComputeV1()
		require.NoError(s.t, err)

		target := netName.WithPrefixForTags("compute")

		key := s.createTagKey("my-key")
		val := s.createTagValue(key.ShortName, "val-A")
		binding := s.createTagBinding(target, val.Name)

		// list tags

		it := s.mock.ListEffectiveTags(s.ctx, &resourcemanagerpb.ListEffectiveTagsRequest{
			Parent: target,
		}).All()
		effectiveTags, err := IteratorToSlice(it)
		require.NoError(s.t, err)
		require.Len(s.t, effectiveTags, 1)
		require.Equal(s.t, fmt.Sprintf("%s/my-key", s.mock.ProjectId()), effectiveTags[0].NamespacedTagKey)
		require.Equal(s.t, fmt.Sprintf("%s/my-key/val-A", s.mock.ProjectId()), effectiveTags[0].NamespacedTagValue)

		// delete

		s.deleteTagBinding(binding.Name)
		s.deleteTagValue(val.Name)
		s.deleteTagKey(key.Name)
	})
}
