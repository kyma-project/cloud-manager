package util

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNames(t *testing.T) {

	t.Run("ParseNameDetail", func(t *testing.T) {

		testCases := []struct {
			path             string
			defns            []*namePartDefn
			values           []string
			rt               ResourceType
			projectId        string
			locationRegionId string
			resourceId       string
		}{
			{
				"projects/my-project/global/networks/my-network",
				[]*namePartDefn{&nameDefnProject, &nameDefnGlobalNetwork},
				[]string{"my-project", "my-network"},
				ResourceTypeGlobalNetwork,
				"my-project", "", "my-network",
			},
			{
				"projects/my-project/operations/my-operation",
				[]*namePartDefn{&nameDefnProject, &nameDefnOperation},
				[]string{"my-project", "my-operation"},
				ResourceTypeGlobalOperation,
				"my-project", "", "my-operation",
			},
			{
				"projects/my-project/regions/my-region/operations/my-operation",
				[]*namePartDefn{&nameDefnProject, &nameDefnRegion, &nameDefnOperation},
				[]string{"my-project", "my-region", "my-operation"},
				ResourceTypeRegionalOperation,
				"my-project", "my-region", "my-operation",
			},
			{
				"projects/my-project/locations/my-location",
				[]*namePartDefn{&nameDefnProject, &nameDefnLocation},
				[]string{"my-project", "my-location"},
				ResourceTypeLocation,
				"my-project", "my-location", "my-location",
			},
			{
				"projects/my-project/locations/my-location/instances/my-instance",
				[]*namePartDefn{&nameDefnProject, &nameDefnLocation, &nameDefnInstance},
				[]string{"my-project", "my-location", "my-instance"},
				ResourceTypeInstance,
				"my-project", "my-location", "my-instance",
			},
			{
				"projects/my-project/locations/my-location/backups/my-backup",
				[]*namePartDefn{&nameDefnProject, &nameDefnLocation, &nameDefnBackup},
				[]string{"my-project", "my-location", "my-backup"},
				ResourceTypeBackup,
				"my-project", "my-location", "my-backup",
			},
			{
				"projects/my-project/locations/my-location/clusters/my-cluster",
				[]*namePartDefn{&nameDefnProject, &nameDefnLocation, &nameDefnCluster},
				[]string{"my-project", "my-location", "my-cluster"},
				ResourceTypeCluster,
				"my-project", "my-location", "my-cluster",
			},
			{
				"projects/my-project/address/my-address",
				[]*namePartDefn{&nameDefnProject, &nameDefnAddress},
				[]string{"my-project", "my-address"},
				ResourceTypeGlobalAddress,
				"my-project", "", "my-address",
			},
			{
				"projects/my-project/services/my-service",
				[]*namePartDefn{&nameDefnProject, &nameDefnService},
				[]string{"my-project", "my-service"},
				ResourceTypeService,
				"my-project", "", "my-service",
			},
			{
				"projects/my-project/regions/my-region",
				[]*namePartDefn{&nameDefnProject, &nameDefnRegion},
				[]string{"my-project", "my-region"},
				ResourceTypeRegion,
				"my-project", "my-region", "my-region",
			},
			{
				"projects/my-project/regions/my-region/routers/my-router",
				[]*namePartDefn{&nameDefnProject, &nameDefnRegion, &nameDefnRouter},
				[]string{"my-project", "my-region", "my-router"},
				ResourceTypeRouter,
				"my-project", "my-region", "my-router",
			},
			{
				"projects/my-project/regions/my-region/addresses/my-address",
				[]*namePartDefn{&nameDefnProject, &nameDefnRegion, &nameDefnAddresses},
				[]string{"my-project", "my-region", "my-address"},
				ResourceTypeRegionalAddress,
				"my-project", "my-region", "my-address",
			},
			{
				"projects/my-project/regions/my-region/subnetworks/my-subnetwork",
				[]*namePartDefn{&nameDefnProject, &nameDefnRegion, &nameDefnSubnetwork},
				[]string{"my-project", "my-region", "my-subnetwork"},
				ResourceTypeSubnetwork,
				"my-project", "my-region", "my-subnetwork",
			},

			// with prefix
			{
				"https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network",
				[]*namePartDefn{&nameDefnProject, &nameDefnGlobalNetwork},
				[]string{"my-project", "my-network"},
				ResourceTypeGlobalNetwork,
				"my-project", "", "my-network",
			},
			{
				"https://example.com/something/projects/my-project/global/networks/my-network",
				[]*namePartDefn{&nameDefnProject, &nameDefnGlobalNetwork},
				[]string{"my-project", "my-network"},
				ResourceTypeGlobalNetwork,
				"my-project", "", "my-network",
			},

			// negative test cases
			{
				// syntax error: missing network id
				"projects/my-project/global/networks",
				nil, nil, ResourceType(""), "", "", "",
			},
			{
				// sequence error: project and network are in wrong order
				"global/networks/my-network/projects/my-project",
				nil, nil, ResourceType(""), "", "", "",
			},
		}

		for _, data := range testCases {
			t.Run(data.path, func(t *testing.T) {
				nd, err := ParseNameDetail(data.path)
				if data.defns == nil {
					// if defns is nil, we expect an error
					assert.Error(t, err)
					return
				}
				require.NoError(t, err)
				for _, defn := range data.defns {
					assert.NotNil(t, defn, "defn must not be nil")
				}
				for _, part := range nd.parts {
					assert.NotNil(t, part.defn)
				}
				var expectedPaths []string
				for i := range data.defns {
					expectedPaths = append(expectedPaths, data.defns[i].Value(data.values[i]).String())
				}
				assert.Equal(t, strings.Join(expectedPaths, "/"), nd.String())
				assert.Equal(t, len(data.defns), len(nd.parts))
				for i := range data.defns {
					assert.Equal(t, data.defns[i].format, nd.parts[i].defn.format)
				}
				for i, val := range data.values {
					assert.Equal(t, val, nd.parts[i].value)
				}
				assert.Equal(t, data.rt, nd.ResourceType())
				assert.Equal(t, data.projectId, nd.ProjectId())
				assert.Equal(t, data.locationRegionId, nd.LocationRegionId())
				assert.Equal(t, data.resourceId, nd.ResourceId())
			})
		}
	})

	t.Run("Constructor", func(t *testing.T) {
		testCases := []struct {
			fn       any
			args     []string
			expected string
		}{
			{NewGlobalNetworkName, []string{"my-project", "my-network"}, "projects/my-project/global/networks/my-network"},
			{NewLocationName, []string{"my-project", "my-location"}, "projects/my-project/locations/my-location"},
			{NewInstanceName, []string{"my-project", "my-location", "my-instance"}, "projects/my-project/locations/my-location/instances/my-instance"},
			{NewBackupName, []string{"my-project", "my-location", "my-backup"}, "projects/my-project/locations/my-location/backups/my-backup"},
			{NewClusterName, []string{"my-project", "my-location", "my-cluster"}, "projects/my-project/locations/my-location/clusters/my-cluster"},
			{NewGlobalAddressName, []string{"my-project", "my-address"}, "projects/my-project/address/my-address"},
			{NewServiceName, []string{"my-project", "my-service"}, "projects/my-project/services/my-service"},
			{NewRegionName, []string{"my-project", "my-region"}, "projects/my-project/regions/my-region"},
			{NewRouterName, []string{"my-project", "my-region", "my-router"}, "projects/my-project/regions/my-region/routers/my-router"},
			{NewRegionalAddressName, []string{"my-project", "my-region", "my-address"}, "projects/my-project/regions/my-region/addresses/my-address"},
			{NewSubnetworkName, []string{"my-project", "my-region", "my-subnetwork"}, "projects/my-project/regions/my-region/subnetworks/my-subnetwork"},
		}

		for _, tc := range testCases {
			fullName := runtime.FuncForPC(reflect.ValueOf(tc.fn).Pointer()).Name()
			parts := strings.Split(fullName, ".")
			lastPart := parts[len(parts)-1]
			t.Run(lastPart, func(t *testing.T) {
				var vals []reflect.Value
				for _, arg := range tc.args {
					vals = append(vals, reflect.ValueOf(arg))
				}
				results := reflect.ValueOf(tc.fn).Call(vals)
				require.Len(t, results, 1)
				nd := results[0].Interface().(NameDetail)
				assert.Equal(t, tc.expected, nd.String())
			})
		}
	})

	t.Run("WithPrefix", func(t *testing.T) {
		name := NewInstanceName("my-project", "my-location", "my-instance")
		assert.Equal(t, "projects/my-project/locations/my-location/instances/my-instance", name.String())
		nameWithPrefix := name.PrefixWithGoogleApisComputeV1()
		assert.Equal(t, prefixGoogleApisComputeV1+"projects/my-project/locations/my-location/instances/my-instance", nameWithPrefix)
	})
}
