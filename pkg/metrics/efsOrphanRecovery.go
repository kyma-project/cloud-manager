package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// EFS orphan-recovery metrics. Disposable alongside the recovery garbage collector.
// All gauges are labeled by shoot/region/mode so a single binary can only ever report for the
// one configured shoot. `remaining_tagged` is the operator's convergence signal.
var (
	efsRecoveryLabels = []string{"shoot", "region", "mode"}

	EfsRecoveryCandidates = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloud_manager_efs_recovery_candidates",
		Help: "Number of in-scope (shoot-tagged) EFS file systems considered by the recovery GC",
	}, efsRecoveryLabels)

	EfsRecoveryKeepers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloud_manager_efs_recovery_keepers",
		Help: "Number of keeper EFS (joined to an NfsInstance status.id)",
	}, efsRecoveryLabels)

	EfsRecoveryOrphans = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloud_manager_efs_recovery_orphans",
		Help: "Number of orphan EFS (in scope, not any NfsInstance status.id)",
	}, efsRecoveryLabels)

	EfsRecoveryAmbiguous = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloud_manager_efs_recovery_ambiguous",
		Help: "Number of ambiguous EFS kept and logged rather than deleted",
	}, efsRecoveryLabels)

	EfsRecoveryRemainingTagged = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloud_manager_efs_recovery_remaining_tagged",
		Help: "Number of shoot-tagged EFS remaining; recovery is converged when this equals keepers",
	}, efsRecoveryLabels)

	EfsRecoveryDeleted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_efs_recovery_deleted_total",
		Help: "Total EFS file systems the recovery GC issued delete for",
	}, efsRecoveryLabels)

	EfsRecoveryMountTargetsDeleted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_efs_recovery_mount_targets_deleted_total",
		Help: "Total EFS mount targets the recovery GC issued delete for",
	}, efsRecoveryLabels)

	EfsRecoverySkipped = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_efs_recovery_skipped_total",
		Help: "Total delete attempts skipped (became keeper, in use, etc.)",
	}, efsRecoveryLabels)

	EfsRecoveryThrottled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_efs_recovery_throttled_total",
		Help: "Total throttling responses encountered by the recovery GC",
	}, efsRecoveryLabels)
)

func init() {
	metrics.Registry.MustRegister(
		EfsRecoveryCandidates,
		EfsRecoveryKeepers,
		EfsRecoveryOrphans,
		EfsRecoveryAmbiguous,
		EfsRecoveryRemainingTagged,
		EfsRecoveryDeleted,
		EfsRecoveryMountTargetsDeleted,
		EfsRecoverySkipped,
		EfsRecoveryThrottled,
	)
}
