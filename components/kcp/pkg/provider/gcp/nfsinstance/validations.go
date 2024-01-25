package nfsinstance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
)

type GcpFileTierValidation interface {
	IsValidCapacity(capacityGb int) (bool, error)
	CanScaleDown() bool
	IsValidNwMask(cidr string) (bool, error)
}

func IsValidCapacity(tier v1beta1.GcpFileTier, capacityGb int) (bool, error) {
	switch tier {
	case v1beta1.BASIC_HDD, v1beta1.STANDARD:
		if capacityGb < 1024 {
			return false, errors.New("Capacity should be > 1 TB")
		} else if capacityGb > 65433 {
			return false, errors.New("Capacity should be < 63.9 TB")
		}
	case v1beta1.BASIC_SSD, v1beta1.PREMIUM:
		if capacityGb < 2560 {
			return false, errors.New("Capacity should be > 2.5 TB")
		} else if capacityGb > 65433 {
			return false, errors.New("Capacity should be < 63.9 TB")
		}
	case v1beta1.ZONAL, v1beta1.ENTERPRISE, v1beta1.REGIONAL:
		if capacityGb < 1024 {
			return false, errors.New("Capacity should be > 1 TB")
		} else if capacityGb > 10240 {
			return false, errors.New("Capacity should be < 10 TB")
		} else if capacityGb%256 != 0 {
			return false, errors.New("Capacity should be in increments of 256 GBs")
		}
	case v1beta1.HIGH_SCALE_SSD:
		if capacityGb < 10240 {
			return false, errors.New("Capacity should be > 10 TB")
		} else if capacityGb > 102400 {
			return false, errors.New("Capacity should be < 100 TB")
		} else if capacityGb%2560 != 0 {
			return false, errors.New("Capacity should be in increments of 2560 GBs")
		}
	default:
		return false, errors.New("Unknown Tier")
	}
	return true, nil
}

func CanScaleDown(tier v1beta1.GcpFileTier) bool {
	return tier == v1beta1.ZONAL || tier == v1beta1.HIGH_SCALE_SSD ||
		tier == v1beta1.ENTERPRISE || tier == v1beta1.REGIONAL
}

func IsValidNwMask(tier v1beta1.GcpFileTier, cidr string) (bool, error) {
	if cidr == "" {
		return true, nil
	}
	switch tier {
	case v1beta1.BASIC_HDD, v1beta1.STANDARD, v1beta1.BASIC_SSD, v1beta1.PREMIUM, v1beta1.ZONAL:
		return checkCidrSuffix(tier, cidr, "/29")
	case v1beta1.ENTERPRISE, v1beta1.REGIONAL:
		return checkCidrSuffix(tier, cidr, "/26")
	case v1beta1.HIGH_SCALE_SSD:
		return checkCidrSuffix(tier, cidr, "/24")
	default:
		return false, errors.New("Unknown Tier")
	}
}

func checkCidrSuffix(tier v1beta1.GcpFileTier, cidr, suffix string) (bool, error) {
	valid := strings.HasSuffix(cidr, suffix)
	if valid {
		return valid, nil
	} else {
		return valid, errors.New(fmt.Sprintf("CIDR block should be %s for Tier: %s", suffix, string(tier)))
	}
}
