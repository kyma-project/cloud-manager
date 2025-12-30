package v2

import (
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

// convertTier converts the CRD tier string to GCP protobuf tier enum.
func convertTier(tier v1beta1.GcpFileTier) filestorepb.Instance_Tier {
	switch tier {
	case v1beta1.BASIC_HDD:
		return filestorepb.Instance_BASIC_HDD
	case v1beta1.BASIC_SSD:
		return filestorepb.Instance_BASIC_SSD
	case v1beta1.ZONAL:
		return filestorepb.Instance_ZONAL
	case v1beta1.REGIONAL:
		return filestorepb.Instance_REGIONAL
	default:
		return filestorepb.Instance_TIER_UNSPECIFIED
	}
}
