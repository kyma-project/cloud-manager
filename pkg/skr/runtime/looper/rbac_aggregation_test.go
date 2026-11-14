package looper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

const skrCRDGroupSuffix = "cloud-resources.kyma-project.io_"

func TestAggregationRolesContract(t *testing.T) {
	view := mustLoadClusterRole(t, "../../../../config/rbac/cloud-manager_view_role.yaml")
	edit := mustLoadClusterRole(t, "../../../../config/rbac/cloud-manager_edit_role.yaml")

	t.Run("view role is read-only and aggregated", func(t *testing.T) {
		require.Equal(t, "kyma-cloud-manager-view", view.Name)
		require.Equal(t, "true", view.Labels["rbac.authorization.k8s.io/aggregate-to-view"])

		for _, rule := range view.Rules {
			assert.NotContains(t, rule.Verbs, "create")
			assert.NotContains(t, rule.Verbs, "update")
			assert.NotContains(t, rule.Verbs, "patch")
			assert.NotContains(t, rule.Verbs, "delete")
			assert.NotContains(t, rule.Verbs, "deletecollection")
		}
	})

	t.Run("view role does not grant access to sensitive resources", func(t *testing.T) {
		for _, rule := range view.Rules {
			for _, apiGroup := range rule.APIGroups {
				assert.Equal(t, "cloud-resources.kyma-project.io", apiGroup,
					"view role must only aggregate cloud-manager CRDs, never core/other groups")
			}
			assert.NotContains(t, rule.Resources, "secrets")
			assert.NotContains(t, rule.Resources, "*")
		}
	})

	t.Run("edit role is mutating and aggregated", func(t *testing.T) {
		require.Equal(t, "kyma-cloud-manager-edit", edit.Name)
		require.Equal(t, "true", edit.Labels["rbac.authorization.k8s.io/aggregate-to-edit"])

		verbs := collectVerbs(edit.Rules)
		assert.Contains(t, verbs, "create")
		assert.Contains(t, verbs, "update")
		assert.Contains(t, verbs, "patch")
		assert.Contains(t, verbs, "delete")
		assert.Contains(t, verbs, "deletecollection")
	})

	t.Run("edit role never uses wildcard resources or verbs", func(t *testing.T) {
		for _, rule := range edit.Rules {
			assert.NotContains(t, rule.Resources, "*")
			assert.NotContains(t, rule.Verbs, "*")
		}
	})

	t.Run("view and edit cover the same resources", func(t *testing.T) {
		assert.Equal(t, collectResources(view.Rules), collectResources(edit.Rules),
			"view and edit must cover identical CRDs so neither drifts behind the other")
	})

	t.Run("roles cover every SKR CRD", func(t *testing.T) {
		covered := collectResources(view.Rules)
		for _, plural := range skrCRDPlurals(t) {
			assert.Contains(t, covered, plural,
				"aggregated roles are missing SKR CRD %q — add it to view and edit roles", plural)
		}
	})
}

// skrCRDPlurals derives the plural name of every SKR (cloud-resources group) CRD
// from the generated CRD base filenames, so a newly added CRD that is not wired
// into the aggregated roles fails this test instead of silently losing coverage.
func skrCRDPlurals(t *testing.T) []string {
	t.Helper()
	dir, err := filepath.Abs("../../../../config/crd/bases")
	require.NoError(t, err)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	var plurals []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, skrCRDGroupSuffix) || !strings.HasSuffix(name, ".yaml") {
			continue
		}
		plural := strings.TrimSuffix(strings.TrimPrefix(name, skrCRDGroupSuffix), ".yaml")
		plurals = append(plurals, plural)
	}
	require.NotEmpty(t, plurals, "found no SKR CRD bases — path or naming changed")
	return plurals
}

func collectResources(rules []rbacv1.PolicyRule) map[string]bool {
	resources := make(map[string]bool)
	for _, rule := range rules {
		for _, r := range rule.Resources {
			resources[r] = true
		}
	}
	return resources
}

func collectVerbs(rules []rbacv1.PolicyRule) map[string]bool {
	verbs := make(map[string]bool)
	for _, rule := range rules {
		for _, verb := range rule.Verbs {
			verbs[verb] = true
		}
	}
	return verbs
}

func mustLoadClusterRole(t *testing.T, relPath string) *rbacv1.ClusterRole {
	t.Helper()
	absPath, err := filepath.Abs(relPath)
	require.NoError(t, err)

	data, err := os.ReadFile(absPath)
	require.NoError(t, err)

	var role rbacv1.ClusterRole
	require.NoError(t, yaml.Unmarshal(data, &role))
	return &role
}
