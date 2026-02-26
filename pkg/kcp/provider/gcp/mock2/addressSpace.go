package mock2

import (
	"fmt"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"k8s.io/utils/ptr"
)

func newAddressSpace() *AddressSpace {
	return &AddressSpace{
		AddressSpace: allocate.MustNewAddressSpace(allocate.PrivateRanges[:]...),
	}
}

type AddressSpace struct {
	allocate.AddressSpace
}

func (as *AddressSpace) Clone() *AddressSpace {
	return &AddressSpace{
		AddressSpace: as.AddressSpace.Clone(),
	}
}

func (as *AddressSpace) AddSubnet(subnet *computepb.Subnetwork) ([]string, error) {
	var added []string
	ar := ptr.Deref(subnet.IpCidrRange, "")
	if ar == "" {
		return added, gcpmeta.NewBadRequestError("subnet ip cidr range is required")
	}
	if err := as.Reserve(ar); err != nil {
		return added, fmt.Errorf("invalid subnet ip cidr range: %w", err)
	}
	added = append(added, ar)

	for _, secondaryRange := range subnet.SecondaryIpRanges {
		ar := ptr.Deref(secondaryRange.IpCidrRange, "")
		if err := as.Reserve(ar); err != nil {
			return added, fmt.Errorf("invalid subnet secondary ip cidr range: %w", err)
		}
		added = append(added, ar)
	}

	return added, nil
}
