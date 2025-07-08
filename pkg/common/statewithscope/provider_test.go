package statewithscope

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	aws       = cloudcontrolv1beta1.ProviderAws
	azure     = cloudcontrolv1beta1.ProviderAzure
	gcp       = cloudcontrolv1beta1.ProviderGCP
	openstack = cloudcontrolv1beta1.ProviderOpenStack
)

func TestProviderPredicate(t *testing.T) {
	testData := []struct {
		title     string
		predicate composed.Predicate
		state     composed.State
		expected  bool
	}{
		// Focal =======================
		{
			"Focal aws predicate aws provider",
			AwsProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildFocal(),
			true,
		},
		{
			"Focal aws predicate gcp provider",
			AwsProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildFocal(),
			false,
		},
		{
			"Focal azure predicate azure provider",
			AzureProviderPredicate,
			sb(s().WithProvider(azure).Build()).BuildFocal(),
			true,
		},
		{
			"Focal azure predicate gcp provider",
			AzureProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildFocal(),
			false,
		},
		{
			"Focal gcp predicate gcp provider",
			GcpProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildFocal(),
			true,
		},
		{
			"Focal gcp predicate aws provider",
			GcpProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildFocal(),
			false,
		},
		{
			"Focal openstack predicate openstack provider",
			OpenStackProviderPredicate,
			sb(s().WithProvider(openstack).Build()).BuildFocal(),
			true,
		},
		{
			"Focal openstack predicate aws provider",
			OpenStackProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildFocal(),
			false,
		},

		// ObjAsScope =======================
		{
			"ObjAsScope aws predicate aws provider",
			AwsProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildObjAsScope(),
			true,
		},
		{
			"ObjAsScope aws predicate gcp provider",
			AwsProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildObjAsScope(),
			false,
		},
		{
			"ObjAsScope azure predicate azure provider",
			AzureProviderPredicate,
			sb(s().WithProvider(azure).Build()).BuildObjAsScope(),
			true,
		},
		{
			"ObjAsScope azure predicate gcp provider",
			AzureProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildObjAsScope(),
			false,
		},
		{
			"ObjAsScope gcp predicate gcp provider",
			GcpProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildObjAsScope(),
			true,
		},
		{
			"ObjAsScope gcp predicate aws provider",
			GcpProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildObjAsScope(),
			false,
		},
		{
			"ObjAsScope openstack predicate openstack provider",
			OpenStackProviderPredicate,
			sb(s().WithProvider(openstack).Build()).BuildObjAsScope(),
			true,
		},
		{
			"ObjAsScope openstack predicate aws provider",
			OpenStackProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildObjAsScope(),
			false,
		},

		// Scope =======================
		{
			"Scope aws predicate aws provider",
			AwsProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildScope(),
			true,
		},
		{
			"Scope aws predicate gcp provider",
			AwsProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildScope(),
			false,
		},
		{
			"Scope azure predicate azure provider",
			AzureProviderPredicate,
			sb(s().WithProvider(azure).Build()).BuildScope(),
			true,
		},
		{
			"Scope azure predicate gcp provider",
			AzureProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildScope(),
			false,
		},
		{
			"Scope gcp predicate gcp provider",
			GcpProviderPredicate,
			sb(s().WithProvider(gcp).Build()).BuildScope(),
			true,
		},
		{
			"Scope gcp predicate aws provider",
			GcpProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildScope(),
			false,
		},
		{
			"Scope openstack predicate openstack provider",
			OpenStackProviderPredicate,
			sb(s().WithProvider(openstack).Build()).BuildScope(),
			true,
		},
		{
			"Scope openstack predicate aws provider",
			OpenStackProviderPredicate,
			sb(s().WithProvider(aws).Build()).BuildScope(),
			false,
		},
	}

	for _, tt := range testData {
		t.Run(tt.title, func(t *testing.T) {
			ctx := context.Background()
			actual := tt.predicate(ctx, tt.state)
			assert.Equal(t, tt.expected, actual, tt.title)
		})
	}
}
