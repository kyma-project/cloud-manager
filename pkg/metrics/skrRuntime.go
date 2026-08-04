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
	}, []string{"kymaName", "globalAccountId", "subAccountId", "shootName", "region", "brokerPlanName"})

	// SkrLooperGateConflictTotal counts SkrGate claim conflicts: a worker found the
	// SKR already claimed by the other sleeve. Sustained high values mean the two
	// pools are fighting over the same SKRs (tune pool sizes or the conflict delay).
	SkrLooperGateConflictTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_skr_looper_gate_conflict_total",
		Help: "Total SkrGate claim conflicts (a worker found the SKR already claimed by the other sleeve), per sleeve",
	}, []string{"sleeve"})

	// SkrLooperGateInFlight is the number of SKRs with a live manager (gate claim
	// held) right now. It must never exceed NotificationConcurrency + CyclicConcurrency.
	SkrLooperGateInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cloud_manager_skr_looper_gate_in_flight",
		Help: "Number of SKRs with a live manager (gate claim held) right now; must never exceed NotificationConcurrency + CyclicConcurrency",
	})

	// SkrLooperNotificationReceivedTotal counts runtime-watcher notifications that
	// carried a valid runtime-id and were forwarded to the notification sleeve.
	SkrLooperNotificationReceivedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cloud_manager_skr_looper_notification_received_total",
		Help: "Total runtime-watcher notifications with a valid runtime-id forwarded to the notification sleeve",
	})

	// SkrLooperNotificationDroppedTotal counts runtime-watcher notifications dropped
	// because the payload had a missing or invalid runtime-id.
	SkrLooperNotificationDroppedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cloud_manager_skr_looper_notification_dropped_total",
		Help: "Total runtime-watcher notifications dropped due to a missing or invalid runtime-id",
	})

	// SkrLooperConnectPhaseSeconds measures per-SKR connect time broken down by phase.
	// Phases: create_manager (KCP secret+Scope reads), skr_readiness (SKR readiness check;
	// zero if skipped), installer (manifest apply + status save; zero if skipped),
	// pre_start (scopeProvider + indexers + controllers setup), start (skrManager.Start, ~10s floor).
	// Labels: phase, kyma, timeout (true/false).
	SkrLooperConnectPhaseSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cloud_manager_skr_looper_connect_phase_seconds",
		Help:    "Duration of each SKR connect phase (create_manager, skr_readiness, installer, pre_start, start) per kyma name and timeout outcome",
		Buckets: []float64{1, 2, 5, 10, 20, 30, 60, 120, 300},
	}, []string{"phase", "kyma", "timeout"})

	// SkrLooperConnectTotalSeconds measures the total wall-clock time for one SKR connect.
	// Labels: kyma, timeout (true/false).
	SkrLooperConnectTotalSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cloud_manager_skr_looper_connect_total_seconds",
		Help:    "Total duration of one SKR connect (pre-amble + Start) per kyma name and timeout outcome",
		Buckets: []float64{1, 2, 5, 10, 20, 30, 60, 120, 300},
	}, []string{"kyma", "timeout"})
)

func init() {
	metrics.Registry.MustRegister(
		SkrRuntimeReconcileTotal,
		SkrRuntimeModuleActiveCount,
		SkrLooperGateConflictTotal,
		SkrLooperGateInFlight,
		SkrLooperNotificationReceivedTotal,
		SkrLooperNotificationDroppedTotal,
		SkrLooperConnectPhaseSeconds,
		SkrLooperConnectTotalSeconds,
	)
}
