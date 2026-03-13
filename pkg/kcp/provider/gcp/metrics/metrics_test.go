package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestParseGoogRequestParams(t *testing.T) {
	tests := []struct {
		name            string
		params          string
		expectedRegion  string
		expectedProject string
	}{
		// Standard cases
		{
			name:            "project and region from standard path",
			params:          "parent=projects/my-project/locations/us-central1",
			expectedRegion:  "us-central1",
			expectedProject: "my-project",
		},
		{
			name:            "project and zone converted to region",
			params:          "parent=projects/test-project/zones/europe-west1-b",
			expectedRegion:  "europe-west1",
			expectedProject: "test-project",
		},
		{
			name:            "URL encoded path",
			params:          "parent=projects%2Fmy-project%2Flocations%2Fasia-south1",
			expectedRegion:  "asia-south1",
			expectedProject: "my-project",
		},
		{
			name:            "resource name with region",
			params:          "name=projects/prod-123/regions/us-east1/networks/default",
			expectedRegion:  "us-east1",
			expectedProject: "prod-123",
		},
		{
			name:            "multiple parameters",
			params:          "project=my-project&parent=projects/my-project/locations/us-west1",
			expectedRegion:  "us-west1",
			expectedProject: "my-project",
		},
		{
			name:            "location is a zone",
			params:          "parent=projects/gcp-123/locations/us-central1-a",
			expectedRegion:  "us-central1",
			expectedProject: "gcp-123",
		},
		{
			name:            "URL encoded with special characters",
			params:          "parent=projects%2Ftest-proj-123%2Fregions%2Feurope-west3",
			expectedRegion:  "europe-west3",
			expectedProject: "test-proj-123",
		},
		{
			name:            "no region in path",
			params:          "parent=projects/my-project",
			expectedRegion:  "",
			expectedProject: "my-project",
		},
		{
			name:            "no project in path",
			params:          "parent=locations/us-central1",
			expectedRegion:  "us-central1",
			expectedProject: "",
		},

		// Edge cases ===============
		{
			name:            "empty params",
			params:          "",
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "malformed params",
			params:          "invalid",
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "multiple equals signs",
			params:          "parent=projects/my-project/locations/us-central1=extra",
			expectedRegion:  "us-central1",
			expectedProject: "my-project",
		},
		{
			name:            "params with no value",
			params:          "parent=",
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "only key no equals",
			params:          "parent",
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "malformed URL encoding",
			params:          "parent=projects%2my-project",
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "template variables in path",
			params:          "parent=projects/{project}/regions/{region}",
			expectedRegion:  "",
			expectedProject: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, project := parseGoogRequestParams(tt.params)
			assert.Equal(t, tt.expectedRegion, region, "region mismatch")
			assert.Equal(t, tt.expectedProject, project, "project mismatch")
		})
	}
}

func TestExtractFromGrpcContext(t *testing.T) {
	tests := []struct {
		name            string
		metadata        map[string]string
		expectedRegion  string
		expectedProject string
	}{
		{
			name: "valid metadata with project and region",
			metadata: map[string]string{
				"x-goog-request-params": "parent=projects/my-project/locations/us-central1",
			},
			expectedRegion:  "us-central1",
			expectedProject: "my-project",
		},
		{
			name: "metadata with zone",
			metadata: map[string]string{
				"x-goog-request-params": "parent=projects/test-project/zones/europe-west1-c",
			},
			expectedRegion:  "europe-west1",
			expectedProject: "test-project",
		},
		{
			name: "URL encoded metadata",
			metadata: map[string]string{
				"x-goog-request-params": "name=projects%2Fgcp-proj%2Fregions%2Fasia-east1%2Finstances%2Ftest",
			},
			expectedRegion:  "asia-east1",
			expectedProject: "gcp-proj",
		},
		{
			name:            "empty metadata map",
			metadata:        map[string]string{},
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name: "metadata without x-goog-request-params",
			metadata: map[string]string{
				"authorization": "Bearer token",
			},
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name: "empty x-goog-request-params",
			metadata: map[string]string{
				"x-goog-request-params": "",
			},
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "no outgoing context",
			metadata:        nil,
			expectedRegion:  "",
			expectedProject: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.metadata != nil {
				md := metadata.New(tt.metadata)
				ctx = metadata.NewOutgoingContext(ctx, md)
			}

			region, project := extractFromGrpcContext(ctx)
			assert.Equal(t, tt.expectedRegion, region, "region mismatch")
			assert.Equal(t, tt.expectedProject, project, "project mismatch")
		})
	}
}
func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "compute API with service prefix and global",
			path:     "/compute/v1/projects/sap-sc-learn/global/addresses/cm-a37f16b9-d5db-4e35-a9bc-8aec7ba5bea8",
			expected: "/compute/v1/projects/{id}/global/addresses/{id}",
		},
		{
			name:     "compute API with region and operation",
			path:     "/compute/v1/projects/sap-sc-learn/regions/us-central1/operations/operation-1768556474843-6487e2472bfc3-f1280de8-246e6450",
			expected: "/compute/v1/projects/{id}/regions/{id}/operations/{id}",
		},
		{
			name:     "compute API with subnetworks",
			path:     "/compute/v1/projects/my-project/regions/us-central1/subnetworks/subnet-123",
			expected: "/compute/v1/projects/{id}/regions/{id}/subnetworks/{id}",
		},
		{
			name:     "basic resource with ID",
			path:     "/v1/projects/my-project/instances/instance-123",
			expected: "/v1/projects/{id}/instances/{id}",
		},
		{
			name:     "custom method after resource ID",
			path:     "/v1/projects/my-project/instances/instance-123:restore",
			expected: "/v1/projects/{id}/instances/{id}:restore",
		},
		{
			name:     "custom method addPeering",
			path:     "/v1/projects/my-project/networks/network-456:addPeering",
			expected: "/v1/projects/{id}/networks/{id}:addPeering",
		},
		{
			name:     "nested resources with multiple IDs",
			path:     "/v1/projects/my-project/regions/us-central1/disks/disk-789",
			expected: "/v1/projects/{id}/regions/{id}/disks/{id}",
		},
		{
			name:     "version prefix v1beta1",
			path:     "/v1beta1/projects/test-proj/resources/res-id",
			expected: "/v1beta1/projects/{id}/resources/{id}",
		},
		{
			name:     "version prefix v2",
			path:     "/v2/projects/abc-123/items/item-xyz",
			expected: "/v2/projects/{id}/items/{id}",
		},
		{
			name:     "path without IDs - only collections",
			path:     "/v1/projects/instances",
			expected: "/v1/projects/instances",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "/",
		},
		{
			name:     "operations path with long ID",
			path:     "/v1/projects/my-proj/operations/operation-abc123-def456-ghi789",
			expected: "/v1/projects/{id}/operations/{id}",
		},
		{
			name:     "locations instead of regions",
			path:     "/v1/projects/my-project/locations/us-central1/instances/inst-1",
			expected: "/v1/projects/{id}/locations/{id}/instances/{id}",
		},
		{
			name:     "zones path",
			path:     "/v1/projects/p1/zones/us-central1-c/instances/i1",
			expected: "/v1/projects/{id}/zones/{id}/instances/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertZoneToRegion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "zone with single letter suffix a",
			input:    "us-central1-a",
			expected: "us-central1",
		},
		{
			name:     "zone with suffix b",
			input:    "europe-west1-b",
			expected: "europe-west1",
		},
		{
			name:     "zone with suffix c",
			input:    "asia-east1-c",
			expected: "asia-east1",
		},
		{
			name:     "already a region",
			input:    "us-central1",
			expected: "us-central1",
		},
		{
			name:     "zone suffix z",
			input:    "us-east1-z",
			expected: "us-east1",
		},
		{
			name:     "short zone",
			input:    "us-a",
			expected: "us",
		},
		{
			name:     "global location",
			input:    "global",
			expected: "global",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "numeric suffix not converted",
			input:    "us-central1-1",
			expected: "us-central1-1",
		},
		{
			name:     "uppercase not converted",
			input:    "us-central1-A",
			expected: "us-central1-A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertZoneToRegion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRegionAndProjectFromPath(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		expectedRegion  string
		expectedProject string
	}{
		{
			name:            "standard project and region",
			path:            "/v1/projects/my-project/regions/us-central1/instances/inst-1",
			expectedRegion:  "us-central1",
			expectedProject: "my-project",
		},
		{
			name:            "project and zone converted to region",
			path:            "/v1/projects/test-proj/zones/europe-west1-a/disks/disk-1",
			expectedRegion:  "europe-west1",
			expectedProject: "test-proj",
		},
		{
			name:            "project and locations",
			path:            "/v1/projects/gcp-123/locations/asia-east1/instances/i1",
			expectedRegion:  "asia-east1",
			expectedProject: "gcp-123",
		},
		{
			name:            "location is a zone",
			path:            "/v1/projects/proj/locations/us-west1-b/resources/r1",
			expectedRegion:  "us-west1",
			expectedProject: "proj",
		},
		{
			name:            "only project no region",
			path:            "/v1/projects/my-project/instances/inst-1",
			expectedRegion:  "",
			expectedProject: "my-project",
		},
		{
			name:            "only region no project",
			path:            "/v1/regions/us-central1/networks/default",
			expectedRegion:  "us-central1",
			expectedProject: "",
		},
		{
			name:            "neither project nor region",
			path:            "/v1/instances/inst-1",
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "template variable project",
			path:            "/v1/projects/{project}/regions/us-central1",
			expectedRegion:  "us-central1",
			expectedProject: "",
		},
		{
			name:            "template variable region",
			path:            "/v1/projects/my-project/regions/{region}",
			expectedRegion:  "",
			expectedProject: "my-project",
		},
		{
			name:            "empty path",
			path:            "",
			expectedRegion:  "",
			expectedProject: "",
		},
		{
			name:            "compute API path",
			path:            "/compute/v1/projects/sap-sc-learn/regions/us-central1/subnetworks/subnet-1",
			expectedRegion:  "us-central1",
			expectedProject: "sap-sc-learn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, project := extractRegionAndProjectFromPath(tt.path)
			assert.Equal(t, tt.expectedRegion, region, "region mismatch")
			assert.Equal(t, tt.expectedProject, project, "project mismatch")
		})
	}
}
