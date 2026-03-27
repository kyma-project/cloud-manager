package looper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

func TestAggregationRolesContract(t *testing.T) {
	view := mustLoadClusterRole(t, "../../../../config/rbac/cloud-manager_view_role.yaml")
	edit := mustLoadClusterRole(t, "../../../../config/rbac/cloud-manager_edit_role.yaml")
	admin := mustLoadClusterRole(t, "../../../../config/rbac/cloud-manager_admin_role.yaml")

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

	t.Run("admin role is explicit and wildcard", func(t *testing.T) {
		require.Equal(t, "kyma-cloud-manager-admin", admin.Name)
		require.Equal(t, "true", admin.Labels["rbac.authorization.k8s.io/aggregate-to-admin"])
		require.Len(t, admin.Rules, 1)
		assert.Equal(t, []string{"*"}, admin.Rules[0].Resources)
		assert.Equal(t, []string{"*"}, admin.Rules[0].Verbs)
	})
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
