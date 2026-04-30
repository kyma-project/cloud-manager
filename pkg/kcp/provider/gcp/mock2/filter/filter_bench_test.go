package filter

import (
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
)

func BenchmarkFilterOnProtobufOneExpression(b *testing.B) {
	obj := &computepb.Network{
		Name:     new("test-network"),
		SelfLink: new("https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"),
		Subnetworks: []string{
			gcputil.NewSubnetworkName("test-project", "test-region", "test-subnetwork-1").String(),
			gcputil.NewSubnetworkName("test-project", "test-region", "test-subnetwork-2").String(),
		},
	}
	fe, _ := NewFilterEngine[*computepb.Network]()
	for i := 0; i < b.N; i++ {
		_, _ = fe.Match(`name = "test-network"`, obj)
	}
}
