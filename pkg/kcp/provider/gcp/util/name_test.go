package util

import (
	"fmt"
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
			//{
			//	"projects/my-project",
			//	[]*namePartDefn{&nameDefnProject},
			//	[]string{"my-project"},
			//	ResourceTypeProject,
			//	"my-project", "", "my-project",
			//},
			//{
			//	"projects/my-project/global/networks/my-network",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnGlobalNetwork},
			//	[]string{"my-project", "my-network"},
			//	ResourceTypeGlobalNetwork,
			//	"my-project", "", "my-network",
			//},
			//{
			//	"projects/my-project/global/operations/my-operation",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnGlobalOperation},
			//	[]string{"my-project", "my-operation"},
			//	ResourceTypeGlobalOperation,
			//	"my-project", "", "my-operation",
			//},
			//{
			//	"projects/my-project/operations/my-operation",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnOperation},
			//	[]string{"my-project", "my-operation"},
			//	ResourceTypeOperation,
			//	"my-project", "", "my-operation",
			//},
			//{
			//	"projects/my-project/regions/my-region/operations/my-operation",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnRegion, &nameDefnOperation},
			//	[]string{"my-project", "my-region", "my-operation"},
			//	ResourceTypeLocationalOperation,
			//	"my-project", "my-region", "my-operation",
			//},
			//{
			//	"projects/my-project/locations/my-location",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnLocation},
			//	[]string{"my-project", "my-location"},
			//	ResourceTypeLocation,
			//	"my-project", "my-location", "my-location",
			//},
			//{
			//	"projects/my-project/locations/my-location/instances/my-instance",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnLocation, &nameDefnInstance},
			//	[]string{"my-project", "my-location", "my-instance"},
			//	ResourceTypeInstance,
			//	"my-project", "my-location", "my-instance",
			//},
			//{
			//	"projects/my-project/locations/my-location/backups/my-backup",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnLocation, &nameDefnBackup},
			//	[]string{"my-project", "my-location", "my-backup"},
			//	ResourceTypeBackup,
			//	"my-project", "my-location", "my-backup",
			//},
			//{
			//	"projects/my-project/locations/my-location/clusters/my-cluster",
			//	[]*namePartDefn{&nameDefnProject, &nameDefnLocation, &nameDefnCluster},
			//	[]string{"my-project", "my-location", "my-cluster"},
			//	ResourceTypeCluster,
			//	"my-project", "my-location", "my-cluster",
			//},
			{
				"projects/my-project/locations/my-location/serviceConnectionPolicies/my-policy",
				[]*namePartDefn{&nameDefnProject, &nameDefnLocation, &nameDefnServiceConnectionPolicy},
				[]string{"my-project", "my-location", "my-policy"},
				ResourceTypeServiceConnectionPolicy,
				"my-project", "my-location", "my-policy",
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
			fn             any
			args           []string
			expectedString string
			expectedType   ResourceType
		}{
			{NewGlobalNetworkName, []string{"my-project", "my-network"}, "projects/my-project/global/networks/my-network", ResourceTypeGlobalNetwork},
			{NewLocationName, []string{"my-project", "my-location"}, "projects/my-project/locations/my-location", ResourceTypeLocation},
			{NewInstanceName, []string{"my-project", "my-location", "my-instance"}, "projects/my-project/locations/my-location/instances/my-instance", ResourceTypeInstance},
			{NewBackupName, []string{"my-project", "my-location", "my-backup"}, "projects/my-project/locations/my-location/backups/my-backup", ResourceTypeBackup},
			{NewClusterName, []string{"my-project", "my-location", "my-cluster"}, "projects/my-project/locations/my-location/clusters/my-cluster", ResourceTypeCluster},
			{ NewServiceConnectionPolicyName, []string{"my-project", "my-location", "my-policy"}, "projects/my-project/locations/my-location/serviceConnectionPolicies/my-policy", ResourceTypeServiceConnectionPolicy},
			{NewGlobalAddressName, []string{"my-project", "my-address"}, "projects/my-project/address/my-address", ResourceTypeGlobalAddress},
			{NewServiceName, []string{"my-project", "my-service"}, "projects/my-project/services/my-service", ResourceTypeService},
			{NewRegionName, []string{"my-project", "my-region"}, "projects/my-project/regions/my-region", ResourceTypeRegion},
			{NewRouterName, []string{"my-project", "my-region", "my-router"}, "projects/my-project/regions/my-region/routers/my-router", ResourceTypeRouter},
			{NewRegionalAddressName, []string{"my-project", "my-region", "my-address"}, "projects/my-project/regions/my-region/addresses/my-address", ResourceTypeRegionalAddress},
			{NewSubnetworkName, []string{"my-project", "my-region", "my-subnetwork"}, "projects/my-project/regions/my-region/subnetworks/my-subnetwork", ResourceTypeSubnetwork},
			{NewLocationalOperationName, []string{"my-project", "my-location", "my-operation"}, "projects/my-project/locations/my-location/operations/my-operation", ResourceTypeLocationalOperation},
			{NewOperationName, []string{"my-project", "my-operation"}, "projects/my-project/operations/my-operation", ResourceTypeOperation},
			{NewGlobalOperationName, []string{"my-project", "my-operation"}, "projects/my-project/global/operations/my-operation", ResourceTypeGlobalOperation},
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
				assert.Equal(t, tc.expectedString, nd.String())
				assert.Equal(t, tc.expectedType, nd.ResourceType())
			})
		}
	})

	t.Run("WithPrefix", func(t *testing.T) {
		name := NewInstanceName("my-project", "my-location", "my-instance")
		assert.Equal(t, "projects/my-project/locations/my-location/instances/my-instance", name.String())
		nameWithPrefix := name.PrefixWithGoogleApisComputeV1()
		assert.Equal(t, prefixGoogleApisComputeV1+"projects/my-project/locations/my-location/instances/my-instance", nameWithPrefix)
	})

	t.Run("Equal", func(t *testing.T) {
		testCases := []struct {
			a        string
			b        string
			expected bool
		}{
			{"projects/my-project/global/networks/my-network", "projects/my-project/global/networks/my-network", true},
			{"projects/my-project/global/networks/my-network-1", "projects/my-project/global/networks/my-network-2", false},
			{"projects/my-project-1/global/networks/my-network", "projects/my-project-2/global/networks/my-network", false},

			{"https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network", "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network", true},
			{"https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network-1", "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network-2", false},
		}

		for i, tc := range testCases {
			t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
				aNd, err := ParseNameDetail(tc.a)
				assert.NoError(t, err)
				bNd, err := ParseNameDetail(tc.b)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, aNd.Equal(bNd))
			})
		}
	})

	t.Run("StartsWith", func(t *testing.T) {
		testCases := []struct {
			prefix   string
			nd       string
			expected bool
		}{
			// starts with project
			{"projects/my-project", "projects/my-project/global/networks/my-network", true},
			{"projects/my-project", "projects/my-project/regions/my-region/operations/my-operation", true},
			{"projects/my-project", "projects/my-project/locations/my-location/instances/my-instance", true},
			{"projects/my-project", "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network", true},
			{"projects/my-project", "https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/addresses/my-address", true},

			// starts with region/location
			{"projects/my-project/regions/my-region", "projects/my-project/regions/my-region/operations/my-operation", true},
			{"projects/my-project/locations/my-location", "projects/my-project/locations/my-location/instances/my-instance", true},
			{"projects/my-project/regions/my-region", "https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/addresses/my-address", true},

			// starts with on equal
			{"projects/my-project/global/networks/my-network", "projects/my-project/global/networks/my-network", true},
			{"projects/my-project/locations/my-location/instances/my-instance", "projects/my-project/locations/my-location/instances/my-instance", true},
			{"projects/my-project/regions/my-region/addresses/my-address", "https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/addresses/my-address", true},
			{"https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/addresses/my-address", "https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/addresses/my-address", true},

			// negative with project
			{"projects/my-project-1", "projects/my-project/global/networks/my-network", false},
			{"projects/my-project-1", "projects/my-project/regions/my-region/operations/my-operation", false},
			{"projects/my-project-1", "projects/my-project/locations/my-location/instances/my-instance", false},
			{"projects/my-project-1", "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-network", false},
			{"projects/my-project-1", "https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/addresses/my-address", false},

			// negative with region/location
			{"projects/my-project/regions/my-region-1", "projects/my-project/regions/my-region/operations/my-operation", false},
			{"projects/my-project/locations/my-location-1", "projects/my-project/locations/my-location/instances/my-instance", false},
			{"projects/my-project/regions/my-region-1", "https://www.googleapis.com/compute/v1/projects/my-project/regions/my-region/addresses/my-address", false},
		}

		for i, tc := range testCases {
			t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
				prefixNd, err := ParseNameDetail(tc.prefix)
				assert.NoError(t, err)
				nd, err := ParseNameDetail(tc.nd)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, nd.StartsWith(prefixNd))
			})
		}
	})
}
