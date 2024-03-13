package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	CloudProviderAWS = "aws"
)

var (
	CloudProviderCallCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_cloud_provider_api_call_total",
		Help: "Total number of cloud provider API calls per provider, method, response code, and region",
	}, []string{"provider", "method", "response_code", "region"})

	CloudProviderCallTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cloud_manager_cloud_provider_api_time_seconds",
		Help:    "Length of cloud provider call time per provider, method, and region",
		Buckets: []float64{0.2, 0.4, 0.6, 0.8, 1.0, 1.5, 2, 3, 4, 5, 7, 9},
	}, []string{"provider", "method", "region"})
)

func init() {
	metrics.Registry.MustRegister(
		CloudProviderCallCount,
		CloudProviderCallTime,
	)
}
