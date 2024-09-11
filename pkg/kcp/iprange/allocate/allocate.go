package allocate

import (
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/common"
)

//const (
//	DefaultNodesCIDR    = "10.250.0.0/22"
//	DefaultPodsCIDR     = "10.96.0.0/13"
//	DefaultServicesCIDR = "10.104.0.0/13"
//)

// https://github.com/kyma-project/kyma-environment-broker/blob/main/internal/networking/cidr.go#L9
var reserved = []string{
	"10.242.0.0/16", "10.64.0.0/11", "10.254.0.0/16", "10.243.0.0/16",
}

const DefaultMaskSize = 22

// AllocateCidr finds an IP range with given maskOnes size such that does not overlap with any
// of the existing ranges. It starts from the first existing range upwards.
func AllocateCidr(maskOnes int, existingRanges []string) (string, error) {
	if len(reserved) == 0 {
		return common.DefaultCloudManagerCidr, nil
	}
	occupied := newRangeList()
	err := occupied.addStrings(existingRanges...)
	if err != nil {
		return "", err
	}
	err = occupied.addStrings(reserved...)
	if err != nil {
		return "", err
	}

	current, _ := parseRange(existingRanges[0])
	current.nextWithOnes(maskOnes)
	for occupied.overlaps(current) {
		current = current.next()
		if current == nil {
			return "", errors.New("unable to find vacant cidr slot")
		}
	}

	return current.s, nil
}
