package util

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

var (
	kymaCRWithThreeModules = NewKymaUnstructured()
	kymaCRWithoutModules   = NewKymaUnstructured()
	kymaCRWithEmptyModules = NewKymaUnstructured()
)

func init() {
	kymaCRWithThreeModules.Object["spec"] = map[string]interface{}{
		"modules": []interface{}{
			map[string]interface{}{
				"name": "foo",
			},
			map[string]interface{}{
				"name": "bar",
			},
			map[string]interface{}{
				"name": "baz",
			},
		},
	}

	kymaCRWithEmptyModules.Object["spec"] = map[string]interface{}{
		"modules": []interface{}{},
	}
}

func Test_NewKymaUnstructured(t *testing.T) {
	k := NewKymaUnstructured()
	assert.IsType(t, &unstructured.Unstructured{}, k)
	assert.Equal(t, "operator.kyma-project.io/v1beta2", k.GetAPIVersion())
	assert.Equal(t, "Kyma", k.GetKind())
}

func Test_NewKymaListUnstructured(t *testing.T) {
	l := NewKymaListUnstructured()
	assert.IsType(t, &unstructured.UnstructuredList{}, l)
	assert.Equal(t, "operator.kyma-project.io/v1beta2", l.GetAPIVersion())
	assert.Equal(t, "KymaList", l.GetKind())
}

func Test_IsKymaModuleListedInSpec(t *testing.T) {
	t.Run("With modules", func(t *testing.T) {
		t.Run("When first", func(t *testing.T) {
			assert.True(t, IsKymaModuleListedInSpec(kymaCRWithThreeModules, "foo"))
		})
		t.Run("When in the middle", func(t *testing.T) {
			assert.True(t, IsKymaModuleListedInSpec(kymaCRWithThreeModules, "bar"))
		})
		t.Run("When last", func(t *testing.T) {
			assert.True(t, IsKymaModuleListedInSpec(kymaCRWithThreeModules, "baz"))
		})
		t.Run("When not listed", func(t *testing.T) {
			assert.False(t, IsKymaModuleListedInSpec(kymaCRWithThreeModules, "non-listed"))
		})
	})

	t.Run("Without modules", func(t *testing.T) {
		assert.False(t, IsKymaModuleListedInSpec(kymaCRWithoutModules, "some"))
	})

	t.Run("With empty modules", func(t *testing.T) {
		assert.False(t, IsKymaModuleListedInSpec(kymaCRWithEmptyModules, "some"))
	})
}

func Test_SetKymaModuleInSpec(t *testing.T) {
	moduleName := "some"
	t.Run("With modules", func(t *testing.T) {
		t.Run("When doesnt exist", func(t *testing.T) {
			k := kymaCRWithThreeModules.DeepCopy()
			assert.NoError(t, SetKymaModuleInSpec(k, moduleName))
			assert.True(t, IsKymaModuleListedInSpec(k, moduleName))
		})

		t.Run("When already exists", func(t *testing.T) {
			k := kymaCRWithThreeModules.DeepCopy()
			assert.NoError(t, SetKymaModuleInSpec(k, "foo"))
			assert.True(t, IsKymaModuleListedInSpec(k, "foo"))
		})
	})

	t.Run("Without modules", func(t *testing.T) {
		k := kymaCRWithoutModules.DeepCopy()
		assert.NoError(t, SetKymaModuleInSpec(k, moduleName))
		assert.True(t, IsKymaModuleListedInSpec(k, moduleName))
	})

	t.Run("With empty modules", func(t *testing.T) {
		k := kymaCRWithEmptyModules.DeepCopy()
		assert.NoError(t, SetKymaModuleInSpec(k, moduleName))
		assert.True(t, IsKymaModuleListedInSpec(k, moduleName))
	})
}

func Test_RemoveKymaModuleFromSpec(t *testing.T) {
	t.Run("With modules", func(t *testing.T) {
		t.Run("Existing first", func(t *testing.T) {
			moduleName := "foo"
			k := kymaCRWithThreeModules.DeepCopy()
			assert.True(t, IsKymaModuleListedInSpec(k, moduleName))
			assert.NoError(t, RemoveKymaModuleFromSpec(k, moduleName))
			assert.False(t, IsKymaModuleListedInSpec(k, moduleName))
		})

		t.Run("Existing in the middle", func(t *testing.T) {
			moduleName := "bar"
			k := kymaCRWithThreeModules.DeepCopy()
			assert.True(t, IsKymaModuleListedInSpec(k, moduleName))
			assert.NoError(t, RemoveKymaModuleFromSpec(k, moduleName))
			assert.False(t, IsKymaModuleListedInSpec(k, moduleName))
		})

		t.Run("Existing last", func(t *testing.T) {
			moduleName := "baz"
			k := kymaCRWithThreeModules.DeepCopy()
			assert.True(t, IsKymaModuleListedInSpec(k, moduleName))
			assert.NoError(t, RemoveKymaModuleFromSpec(k, moduleName))
			assert.False(t, IsKymaModuleListedInSpec(k, moduleName))
		})

		t.Run("Non-existing", func(t *testing.T) {
			moduleName := "some"
			k := kymaCRWithThreeModules.DeepCopy()
			assert.False(t, IsKymaModuleListedInSpec(k, moduleName))
			assert.NoError(t, RemoveKymaModuleFromSpec(k, moduleName))
			assert.False(t, IsKymaModuleListedInSpec(k, moduleName))
		})
	})

	t.Run("Without modules", func(t *testing.T) {
		moduleName := "some"
		k := kymaCRWithoutModules.DeepCopy()
		assert.NoError(t, RemoveKymaModuleFromSpec(k, moduleName))
		assert.False(t, IsKymaModuleListedInSpec(k, moduleName))
	})

	t.Run("With empty modules", func(t *testing.T) {
		moduleName := "some"
		k := kymaCRWithEmptyModules.DeepCopy()
		assert.NoError(t, RemoveKymaModuleFromSpec(k, moduleName))
		assert.False(t, IsKymaModuleListedInSpec(k, moduleName))
	})
}

func Test_StatusE2E(t *testing.T) {
	k := kymaCRWithThreeModules.DeepCopy()
	moduleName := "some"

	assert.Equal(t, KymaModuleStateNotPresent, GetKymaModuleStateFromStatus(k, moduleName))

	assert.NoError(t, SetKymaModuleStateToStatus(k, moduleName, KymaModuleStateReady))
	assert.Equal(t, KymaModuleStateReady, GetKymaModuleStateFromStatus(k, moduleName))

	assert.NoError(t, RemoveKymaModuleStateFromStatus(k, moduleName))
	assert.Equal(t, KymaModuleStateNotPresent, GetKymaModuleStateFromStatus(k, moduleName))
}
