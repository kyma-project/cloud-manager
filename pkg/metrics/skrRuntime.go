package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	SkrRuntimeReconcileTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_skr_runtime_reconcile_total",
		Help: "Total number of SKR reconciliation connections per kyma name",
	}, []string{"kyma"})

	SkrRuntimeModuleActiveCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloud_manager_skr_runtime_module_active_count",
		Help: "Number of SKRs with currently active cloud-manager module per kyma name",
	}, []string{"kyma"})
)

func init() {
	metrics.Registry.MustRegister(
		SkrRuntimeReconcileTotal,
		SkrRuntimeModuleActiveCount,
	)
}
