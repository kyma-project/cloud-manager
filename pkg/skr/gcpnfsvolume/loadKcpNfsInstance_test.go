package gcpnfsvolume

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/zapr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	lkniKyma = "skr"
	lkniNs   = "test"
	lkniVol  = "vol"
)

func lkniInstance(name string) *cloudcontrolv1beta1.NfsInstance {
	return &cloudcontrolv1beta1.NfsInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: lkniNs,
			Labels: map[string]string{
				cloudcontrolv1beta1.LabelKymaName:        lkniKyma,
				cloudcontrolv1beta1.LabelRemoteName:      lkniVol,
				cloudcontrolv1beta1.LabelRemoteNamespace: lkniNs,
			},
		},
	}
}

// lkniState builds a gcpnfsvolume State whose KCP cluster is seeded with the given
// NfsInstances, plus a ctx carrying an observer-backed logger (through the production
// LogFilterSink chain) and the recorder.
func lkniState(t *testing.T, listErr error, instances ...*cloudcontrolv1beta1.NfsInstance) (context.Context, *State, *observer.ObservedLogs) {
	t.Helper()

	kcpBuilder := fake.NewClientBuilder().WithScheme(commonscheme.KcpScheme)
	for _, in := range instances {
		kcpBuilder = kcpBuilder.WithObjects(in)
	}
	if listErr != nil {
		kcpBuilder = kcpBuilder.WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return listErr
			},
		})
	}
	kcpClient := kcpBuilder.Build()
	kcpCluster := composed.NewStateCluster(kcpClient, kcpClient, nil, commonscheme.KcpScheme)

	vol := &cloudresourcesv1beta1.GcpNfsVolume{
		ObjectMeta: metav1.ObjectMeta{Name: lkniVol, Namespace: lkniNs},
	}
	skrClient := fake.NewClientBuilder().WithScheme(commonscheme.SkrScheme).WithObjects(vol).Build()
	skrCluster := composed.NewStateCluster(skrClient, skrClient, nil, commonscheme.SkrScheme)

	baseState := composed.NewStateFactory(skrCluster).NewState(
		types.NamespacedName{Name: lkniVol, Namespace: lkniNs}, vol)

	state := &State{
		State:      baseState,
		KymaRef:    klog.ObjectRef{Name: lkniKyma, Namespace: lkniNs},
		KcpCluster: kcpCluster,
	}

	core, logs := observer.New(zapcore.DebugLevel)
	logger := zapr.NewLogger(zap.New(core, zap.AddCaller()))
	logger = logger.WithSink(util.NewLogFilterSink(logger.GetSink()))
	ctx := composed.LoggerIntoCtx(context.Background(), logger)
	return ctx, state, logs
}

func TestLoadKcpNfsInstanceZeroMatch(t *testing.T) {
	ctx, state, logs := lkniState(t, nil)

	err, _ := loadKcpNfsInstance(ctx, state)

	require.Nil(t, err)
	assert.Nil(t, state.KcpNfsInstance)
	assert.Empty(t, logs.All())
}

func TestLoadKcpNfsInstanceOneMatch(t *testing.T) {
	ctx, state, logs := lkniState(t, nil, lkniInstance("only-instance"))

	err, _ := loadKcpNfsInstance(ctx, state)

	require.Nil(t, err)
	require.NotNil(t, state.KcpNfsInstance)
	assert.Equal(t, "only-instance", state.KcpNfsInstance.Name)
	assert.Empty(t, logs.All())
}

func TestLoadKcpNfsInstanceMultipleMatchWarnsAndSelectsFirst(t *testing.T) {
	ctx, state, logs := lkniState(t, nil, lkniInstance("bbb-instance"), lkniInstance("aaa-instance"))

	err, _ := loadKcpNfsInstance(ctx, state)

	require.Nil(t, err)
	require.NotNil(t, state.KcpNfsInstance)
	assert.Equal(t, "aaa-instance", state.KcpNfsInstance.Name, "must select lexicographically-first")

	entries := logs.All()
	require.Len(t, entries, 1)
	assert.Equal(t, zapcore.WarnLevel, entries[0].Level, "must emit WARNING, not INFO")
	assert.Equal(t, "Found more than one KCP NfsInstance", entries[0].Message)
	names := entries[0].ContextMap()["names"]
	assert.Contains(t, names, "aaa-instance")
	assert.Contains(t, names, "bbb-instance")
}

func TestLoadKcpNfsInstanceListError(t *testing.T) {
	ctx, state, logs := lkniState(t, errors.New("boom"))

	err, _ := loadKcpNfsInstance(ctx, state)

	assert.Equal(t, composed.StopWithRequeue, err)
	entries := logs.All()
	require.Len(t, entries, 1)
	assert.Equal(t, zapcore.ErrorLevel, entries[0].Level)
	assert.Equal(t, "Error loading KCP NfsInstance", entries[0].Message)
}
