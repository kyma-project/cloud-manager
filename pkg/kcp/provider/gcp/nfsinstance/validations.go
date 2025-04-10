package nfsinstance

import (
	"errors"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

func IsValidCapacity(tier v1beta1.GcpFileTier, capacityGb int) (bool, error) {
	switch tier {
	case v1beta1.BASIC_HDD:
		if capacityGb < 1024 {
			return false, errors.New("capacity should be > 1 TB")
		} else if capacityGb > 65433 {
			return false, errors.New("capacity should be < 63.9 TB")
		}
	case v1beta1.BASIC_SSD:
		if capacityGb < 2560 {
			return false, errors.New("capacity should be > 2.5 TB")
		} else if capacityGb > 65433 {
			return false, errors.New("capacity should be < 63.9 TB")
		}
	case v1beta1.ZONAL, v1beta1.REGIONAL:
		if capacityGb < 1024 {
			return false, errors.New("capacity should be > 1 TB")
		} else if capacityGb > 10240 {
			return false, errors.New("capacity should be < 10 TB")
		} else if capacityGb%256 != 0 {
			return false, errors.New("capacity should be in increments of 256 GBs")
		}
	default:
		return false, errors.New("unknown Tier")
	}
	return true, nil
}

func CanScaleDown(tier v1beta1.GcpFileTier) bool {
	return tier == v1beta1.ZONAL || tier == v1beta1.REGIONAL
}
