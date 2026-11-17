package feature

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
)

// cloudResourcesPkgPath is the Go package path for the cloud-resources v1beta1
// API types. Only types from this package are cloud-manager CRDs; other types
// registered under the same GroupVersion (e.g. metav1 options types added by
// metav1.AddToGroupVersion) are excluded from the tagging guard.
const cloudResourcesPkgPath = "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

// TestEverySkrCrdIsTagged enumerates every type registered under the SKR
// cloud-resources GroupVersion and asserts that each root CRD (non-List) kind:
//  1. Implements featuretypes.FeatureAwareObject.
//  2. Returns a SpecificToFeature() value that is neither "" nor FeatureUnknown.
//
// A CRD that fails either check is a security gap under secure-by-default: the
// untagged sentinel routes the resource to the always-on bucket, bypassing
// feature-flag gating. This test fails at CI time instead of silently at
// runtime.
func TestEverySkrCrdIsTagged(t *testing.T) {
	knownTypes := commonscheme.SkrScheme.KnownTypes(cloudresourcesv1beta1.GroupVersion)

	// Collect kinds in deterministic order for stable failure messages.
	// Only include types whose Go package is the cloud-resources v1beta1 package.
	// Types from other packages (e.g. k8s.io metav1 options registered via
	// metav1.AddToGroupVersion) are not CRDs and are excluded.
	kinds := make([]string, 0, len(knownTypes))
	for kind, rt := range knownTypes {
		// Dereference pointer types to reach the struct's package path.
		pkg := rt.PkgPath()
		if pkg == "" && rt.Kind().String() == "ptr" {
			pkg = rt.Elem().PkgPath()
		}
		if pkg != cloudResourcesPkgPath {
			t.Logf("  SKIP %s (package %s, not a cloud-resources CRD)", kind, pkg)
			continue
		}
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)

	type result struct {
		kind    string
		feature featuretypes.FeatureName
		problem string
	}
	var offenders []result

	t.Log("SKR cloud-resources CRD kinds enumerated:")
	for _, kind := range kinds {
		if strings.HasSuffix(kind, "List") {
			t.Logf("  SKIP %s (List type)", kind)
			continue
		}

		gvk := cloudresourcesv1beta1.GroupVersion.WithKind(kind)
		obj, err := commonscheme.SkrScheme.New(gvk)
		if err != nil {
			offenders = append(offenders, result{
				kind:    kind,
				problem: fmt.Sprintf("scheme.New failed: %v", err),
			})
			continue
		}

		fa, ok := obj.(featuretypes.FeatureAwareObject)
		if !ok {
			offenders = append(offenders, result{
				kind:    kind,
				problem: "does not implement FeatureAwareObject",
			})
			t.Logf("  FAIL %s: does not implement FeatureAwareObject", kind)
			continue
		}

		feature := fa.SpecificToFeature()
		t.Logf("  OK   %s: SpecificToFeature()=%q", kind, feature)

		if feature == "" || feature == featuretypes.FeatureUnknown {
			offenders = append(offenders, result{
				kind:    kind,
				feature: feature,
				problem: fmt.Sprintf("SpecificToFeature() returned untagged sentinel %q", feature),
			})
		}
	}

	var msgs []string
	for _, o := range offenders {
		msgs = append(msgs, fmt.Sprintf(
			"CRD %s is not tagged: %s (implement SpecificToFeature returning a named feature or FeatureFoundation)",
			o.kind, o.problem,
		))
	}
	assert.Empty(t, offenders, "untagged SKR CRDs found:\n%s", strings.Join(msgs, "\n"))
}
