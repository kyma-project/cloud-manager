package mock2

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestMockE2E(t *testing.T) {

	srv := New()
	mock := srv.NewSubscription("e2e-01")

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
	defer cancel()

	vop, err := mock.InsertNetwork(ctx, &computepb.InsertNetworkRequest{
		Project: mock.ProjectId(),
		NetworkResource: &computepb.Network{
			Name: ptr.To("test-network"),
		},
	})
	assert.NoError(t, err)

	err = vop.Wait(ctx)
	assert.NoError(t, err)
}
