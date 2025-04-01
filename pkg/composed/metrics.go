package composed

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	Reconcile = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_reconcile",
		Help: "Total number of SKR reconciliation connections per kyma name",
	}, []string{"controller", "name", "result"})
)

func init() {
	metrics.Registry.MustRegister(
		Reconcile,
	)
}
