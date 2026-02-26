package filter

import (
	"encoding/json"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/file/v1"
	"google.golang.org/api/googleapi"
	"k8s.io/utils/ptr"
)

func TestFilter(t *testing.T) {

	t.Run("computepb.Network", func(t *testing.T) {

		fe, err := NewFilterEngine[*computepb.Network]()
		assert.NoError(t, err)

		obj := &computepb.Network{
			Name:     ptr.To("test-network"),
			SelfLink: ptr.To("https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"),
			Subnetworks: []string{
				gcputil.NewSubnetworkName("test-project", "test-region", "test-subnetwork-1").String(),
				gcputil.NewSubnetworkName("test-project", "test-region", "test-subnetwork-2").String(),
			},
		}

		testCases := []struct {
			filter string
			match  bool
			err    string
		}{
			{`name = "test-network"`, true, ""},
			{`name eq test-network`, true, ""},
			{`name = "non-existing-network"`, false, ""},
			{`name eq non-existing-network`, false, ""},
			{`invalid filter`, false, "Syntax error"},

			// AIP-160 explicit AND - match
			{`name = "test-network" AND self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"`, true, ""},
			// AIP-160 explicit AND - miss
			{`name = "foo" AND self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"`, false, ""},
			// AIP-160 explicit AND with parentheses - match
			{`(name = "test-network") AND (self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network")`, true, ""},
			// AIP-160 implicit AND with parentheses - match - TODO: not yet supported, though it's a valid GCP filter
			{`(name = "test-network") (self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network")`, false, "Syntax error"},

			// AIP-160 explicit OR - match, both true
			{`name = "test-network" OR self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"`, true, ""},
			// AIP-160 explicit AND - match, one true, one false
			{`name = "foo" OR self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"`, true, ""},
			// AIP-160 explicit AND - miss, both false
			{`name = "foo" OR self_link = "bar"`, false, ""},
		}

		for _, tc := range testCases {
			t.Run(tc.filter, func(t *testing.T) {
				ok, err := fe.Match(tc.filter, obj)
				if tc.err != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tc.err)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.match, ok)
			})
		}

	})

	t.Run("file.Operation", func(t *testing.T) {

		fe, err := NewFilterEngine[*file.Operation]()
		assert.NoError(t, err)

		metadata := map[string]interface{}{
			"verb":   "restore",
			"target": "projects/test-project/locations/test-location/instances/test-instance",
		}
		metadataBytes, err := json.Marshal(metadata)
		assert.NoError(t, err)

		obj := &file.Operation{
			Name:     "test-operation",
			Done:     true,
			Metadata: googleapi.RawMessage(metadataBytes),
		}

		testCases := []struct {
			filter string
			match  bool
			err    string
		}{
			{`name = "test-operation"`, true, ""},
			{`done = true`, true, ""},
			{`metadata.verb="restore"`, true, ""},
			{`metadata.target="projects/test-project/locations/test-location/instances/test-instance"`, true, ""},
			{`metadata.verb="restore" AND metadata.target="projects/test-project/locations/test-location/instances/test-instance"`, true, ""},

			// negative cases
			{`name = "foo"`, false, ""},
			{`done = false`, false, ""},
			{`metadata.verb="backup"`, false, ""},
			{`metadata.target="something"`, false, ""},
			{`metadata.verb="restore" AND metadata.target="something"`, false, ""},
			{`metadata.verb="backup" AND metadata.target="projects/test-project/locations/test-location/instances/test-instance"`, false, ""},
		}

		for _, tc := range testCases {
			t.Run(tc.filter, func(t *testing.T) {
				ok, err := fe.Match(tc.filter, obj)
				if tc.err != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tc.err)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.match, ok)
			})
		}

	})
}
