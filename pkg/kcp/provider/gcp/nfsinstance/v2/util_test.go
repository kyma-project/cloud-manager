package v2

import (
	"testing"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
)

// TestConvertTier tests tier conversion business logic (CRD enum → GCP protobuf enum)
func TestConvertTier(t *testing.T) {
	tests := []struct {
		name     string
		crdTier  v1beta1.GcpFileTier
		expected filestorepb.Instance_Tier
	}{
		{"BASIC_HDD", v1beta1.BASIC_HDD, filestorepb.Instance_BASIC_HDD},
		{"BASIC_SSD", v1beta1.BASIC_SSD, filestorepb.Instance_BASIC_SSD},
		{"ZONAL", v1beta1.ZONAL, filestorepb.Instance_ZONAL},
		{"REGIONAL", v1beta1.REGIONAL, filestorepb.Instance_REGIONAL},
		{"unknown tier defaults to TIER_UNSPECIFIED", v1beta1.GcpFileTier("UNKNOWN"), filestorepb.Instance_TIER_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTier(tt.crdTier)
			assert.Equal(t, tt.expected, result)
		})
	}
}
