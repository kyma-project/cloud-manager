package filter

import (
	"encoding/json"
	"fmt"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			// partial string match
			{`name:"test"`, true, ""},
			{`name:"network"`, true, ""},
			{`name:"foobarbaz"`, false, ""},

			// list contains
			{`subnetworks:("test-subnetwork-1")`, true, ""},
			{`subnetworks:("foobarbaz")`, false, ""},

			// AIP-160 explicit AND - match
			{`name = "test-network" AND self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"`, true, ""},
			// AIP-160 explicit AND - miss
			{`name = "foo" AND self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network"`, false, ""},
			// AIP-160 explicit AND with parentheses - match
			{`(name = "test-network") AND (self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network")`, true, ""},
			// AIP-160 implicit AND with parentheses - match
			{`(name = "test-network") (self_link = "https://www.googleapis.com/compute/v1/projects/test-project/global/networks/test-network")`, true, ""},

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
					require.Error(t, err, fe.TranslatedExpression())
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

	t.Run("computepb.Address", func(t *testing.T) {
		// (purpose=NAT_AUTO)(users:(routers/shoot--kyma-dev--c-82be022-cloud-router))
		fe, err := NewFilterEngine[*computepb.Address]()
		assert.NoError(t, err)

		obj := &computepb.Address{
			Address:      ptr.To("10.250.0.0"),
			PrefixLength: ptr.To(int32(16)),
			Id:           ptr.To(uint64(1358567327454931553)),
			Kind:         ptr.To("compute#address"),
			Name:         ptr.To("test-address"),
			Region:       ptr.To("https://www.googleapis.com/compute/v1/projects/my-project/regions/europe-west3"),
			SelfLink:     ptr.To("https://www.googleapis.com/compute/v1/projects/my-project/regions/europe-west3/addresses/test-address"),
			Status:       ptr.To("IN_USE"),
			Purpose:      ptr.To("NAT_AUTO"),
			NetworkTier:  ptr.To("PREMIUM"),
			Users: []string{
				"https://www.googleapis.com/compute/v1/projects/my-project/regions/europe-west3/routers/my-router",
			},
		}

		testCases := []struct {
			filter string
			match  bool
		}{
			{`purpose="NAT_AUTO"`, true},
			{`purpose="FOO"`, false},

			{`(purpose="NAT_AUTO")(name="test-address")`, true},
			{`(purpose="NAT_AUTO")(name="foo")`, false},

			{`users:("routers/my-router")`, true},
			{`(users:("routers/my-router"))`, true},
			{`users:"routers/my-router"`, true},
			{`(users:("routers/foo"))`, false},

			{`(users:("routers/my-router"))(purpose="NAT_AUTO")(name="test-address")`, true},
			{`(users:("routers/foo"))(purpose="NAT_AUTO")(name="test-address")`, false},
			{`(users:("routers/my-router"))(purpose="FOO")(name="test-address")`, false},
			{`(users:("routers/my-router"))(purpose="NAT_AUTO")(name="foo")`, false},
		}

		for _, tc := range testCases {
			t.Run(tc.filter, func(t *testing.T) {
				ok, err := fe.Match(tc.filter, obj)
				assert.NoError(t, err)
				assert.Equal(t, tc.match, ok)
			})
		}
	})

	t.Run("filestorepb.Backup with hyphenated label keys (GetSharedBackupsFilter)", func(t *testing.T) {
		fe, err := NewFilterEngine[*filestorepb.Backup]()
		assert.NoError(t, err)

		shootName := "5e32a9dd-4e68-47c7-aac7-64a4880a00d7"

		obj := &filestorepb.Backup{
			Name: "projects/test-project/locations/us-west1/backups/test-backup",
			Labels: map[string]string{
				"managed-by":                          "cloud-manager",
				"scope-name":                          shootName,
				util.GcpLabelShootName:                shootName,
				fmt.Sprintf("cm-allow-%s", shootName): util.GcpLabelBackupAccessibleFrom,
			},
		}

		sharedFilter := gcpclient.GetSharedBackupsFilter(shootName, shootName)
		ok, err := fe.Match(sharedFilter, obj)
		require.NoError(t, err, "GetSharedBackupsFilter should match backup with matching shoot label")
		assert.True(t, ok)

		// Backup without the allow label should NOT match
		objNoLabel := &filestorepb.Backup{
			Name: "projects/test-project/locations/us-west1/backups/other-backup",
			Labels: map[string]string{
				"managed-by": "cloud-manager",
			},
		}
		ok, err = fe.Match(sharedFilter, objNoLabel)
		require.NoError(t, err)
		assert.False(t, ok)
	})

}
