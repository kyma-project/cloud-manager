package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	CloudProviderAWS = "aws"
	CloudProviderGCP = "gcp"
)

var (
	CloudProviderCallCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cloud_manager_cloud_provider_api_call_total",
		Help: "Total number of cloud provider API calls per provider, method, response code, and region",
	}, []string{"provider", "method", "response_code", "region"})
)

func init() {
	metrics.Registry.MustRegister(
		CloudProviderCallCount,
	)
}
