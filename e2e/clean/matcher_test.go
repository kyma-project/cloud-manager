package clean

import (
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type matcherCase struct {
	gvk schema.GroupVersionKind
	expected bool
}

func TestMatcher(t *testing.T) {

	testCases := []struct {
		title string
		matcher Matcher
		cases []matcherCase
	} {
		{
			"whole group w/out one kind",
			MatchAll(MatchingGroup(cloudresourcesv1beta1.GroupVersion.Group), NotMatch(MatchingKind("CloudResources"))),
			[]matcherCase{
				{ cloudresourcesv1beta1.GroupVersion.WithKind("IpRange"), true},
				{ cloudresourcesv1beta1.GroupVersion.WithKind("AwsNfsVolume"), true},
				{ cloudresourcesv1beta1.GroupVersion.WithKind("CloudResources"), false},

				{ infrastructuremanagerv1.GroupVersion.WithKind("Runtime"), false},
				{ infrastructuremanagerv1.GroupVersion.WithKind("GardenerCluster"), false},

				{ schema.ParseGroupKind("pod").WithVersion("v"), false},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			for _, c := range tc.cases {
				t.Run(c.gvk.String(), func(t *testing.T) {
					actual := tc.matcher(c.gvk, nil)
					assert.Equal(t, c.expected, actual)
				})
			}
		})
	}

}
