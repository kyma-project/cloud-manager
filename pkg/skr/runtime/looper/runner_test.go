package looper

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// testNfsBackupObj is a minimal client.Object that reports itself as FeatureNfsBackup.
// It is used to populate registries without registering any real CRD types.
type testNfsBackupObj struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (t *testNfsBackupObj) DeepCopyObject() k8sruntime.Object { c := *t; return &c }
func (t *testNfsBackupObj) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

// fakeSkrManager satisfies skrmanager.SkrManager for the methods runner.go calls.
// Faked because the real one holds a live controller-runtime manager and cluster connection.
// Any unimplemented method from the embedded manager.Manager will panic — a test
// that triggers one of those methods needs a new override.
type fakeSkrManager struct {
	manager.Manager // nil — panics on unimplemented calls
	scheme          *k8sruntime.Scheme
}

func newFakeSkrManager() *fakeSkrManager {
	return &fakeSkrManager{scheme: k8sruntime.NewScheme()}
}

func (f *fakeSkrManager) GetLogger() logr.Logger        { return logr.Discard() }
func (f *fakeSkrManager) GetScheme() *k8sruntime.Scheme { return f.scheme }
func (f *fakeSkrManager) Start(_ context.Context) error { return nil }
func (f *fakeSkrManager) KymaRef() klog.ObjectRef {
	return klog.ObjectRef{Namespace: "default", Name: "kyma-test"}
}
func (f *fakeSkrManager) IgnoreWatchErrors(bool)               {}
func (f *fakeSkrManager) GetFieldIndexer() client.FieldIndexer { return &noopFieldIndexer{} }

// noopFieldIndexer satisfies client.FieldIndexer without interacting with any cache.
type noopFieldIndexer struct{}

func (n *noopFieldIndexer) IndexField(_ context.Context, _ client.Object, _ string, _ client.IndexerFunc) error {
	return nil
}

// noopReconcilerFactory satisfies registry.Builder.WithFactory — never called in tests
// because SetupWithManager is only reached in the Starting path. Faked because the real
// factory wires controllers that need a live controller-runtime manager.
type noopReconcilerFactory struct{}

func (f *noopReconcilerFactory) New(_ reconcile2.ReconcilerArguments) ctrlreconcile.Reconciler {
	return nil
}

// capturingStatusSaver records the SkrStatus pointer passed to Save so tests can
// inspect its handles after Run returns. Faked because the real saver patches to the
// KCP cluster via a live k8s client.
type capturingStatusSaver struct {
	captured *SkrStatus
}

func (s *capturingStatusSaver) Save(_ context.Context, skrStatus *SkrStatus) error {
	s.captured = skrStatus
	return nil
}

// newTestRegistryWithIndexerAndBuilder returns a registry containing one indexer and
// one controller builder — both backed by testNfsBackupObj (FeatureNfsBackup).
func newTestRegistryWithIndexerAndBuilder() registry.SkrRegistry {
	reg := registry.New(k8sruntime.NewScheme())
	reg.IndexField(&testNfsBackupObj{}, ".spec.test", func(_ client.Object) []string { return nil })
	b := reg.Register()
	_ = b.For(&testNfsBackupObj{}).WithFactory(&noopReconcilerFactory{}).Complete()
	return reg
}

// newTestIndexerOnlyRegistry returns a registry with one indexer and no builders.
// Used when we want to exercise the Starting path without reaching SetupWithManager.
func newTestIndexerOnlyRegistry() registry.SkrRegistry {
	reg := registry.New(k8sruntime.NewScheme())
	reg.IndexField(&testNfsBackupObj{}, ".spec.test", func(_ client.Object) []string { return nil })
	return reg
}

// TestRunnerGuardReadsApiEnabled loads the non-parity fixture:
//
//	apiEnabled.defaultRule = disabled  → ApiEnabled.Value() = false
//	apiDisabled.defaultRule = enabled  → ApiDisabled.Value() = false
//
// The legacy guard !apiDisabled would equal !false=true (Starting).
// The new guard  apiEnabled  equals false (ApiNotEnabled).
//
// Recording "ApiNotEnabled" for both the indexer and the controller proves the runner
// reads the positive apiEnabled flag, not the legacy !apiDisabled expression.
func TestRunnerGuardReadsApiEnabled(t *testing.T) {
	require.NoError(t, feature.Initialize(context.Background(), logr.Discard(),
		feature.WithFile("testdata/apiEnabledNonParity.yaml")))
	t.Cleanup(func() {
		if err := feature.Initialize(context.Background(), logr.Discard(),
			feature.WithFile("testdata/ffAllEnabled.yaml")); err != nil {
			t.Errorf("cleanup: restore ffAllEnabled.yaml: %v", err)
		}
	})

	reg := newTestRegistryWithIndexerAndBuilder()
	saver := &capturingStatusSaver{}
	runner := NewSkrRunner(reg, nil, saver, "kyma-test")
	mgr := newFakeSkrManager()

	require.NoError(t, runner.Run(context.Background(), mgr))

	require.Len(t, saver.captured.handles, 2)
	assert.Equal(t, "Indexer", saver.captured.handles[0].title)
	assert.Equal(t, []string{"ApiNotEnabled"}, saver.captured.handles[0].outcomes)
	assert.Equal(t, "Controller", saver.captured.handles[1].title)
	assert.Equal(t, []string{"ApiNotEnabled"}, saver.captured.handles[1].outcomes)
}

// TestRunnerIndexerStartsWhenApiEnabled uses the all-enabled fixture loaded by TestMain
// (apiEnabled.defaultRule = enabled → ApiEnabled.Value() = true).
//
// With the guard passing, the indexer records "Starting" instead of "ApiNotEnabled".
// An indexer-only registry is used so the test never reaches SetupWithManager,
// which requires a real controller-runtime manager.
func TestRunnerIndexerStartsWhenApiEnabled(t *testing.T) {
	// Self-contained: do not rely on TestMain's global provider surviving other tests.
	require.NoError(t, feature.Initialize(context.Background(), logr.Discard(),
		feature.WithFile("testdata/ffAllEnabled.yaml")))

	reg := newTestIndexerOnlyRegistry()
	saver := &capturingStatusSaver{}
	runner := NewSkrRunner(reg, nil, saver, "kyma-test")
	mgr := newFakeSkrManager()

	require.NoError(t, runner.Run(context.Background(), mgr))

	require.Len(t, saver.captured.handles, 1)
	assert.Equal(t, "Indexer", saver.captured.handles[0].title)
	assert.Equal(t, []string{"Starting"}, saver.captured.handles[0].outcomes)
}
