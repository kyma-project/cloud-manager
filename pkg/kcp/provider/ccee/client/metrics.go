package client

import (
	"fmt"
	cceemeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/meta"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	pph "github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type HTTPClient struct {
	http.Client
}

func instrumentCounter(next http.RoundTripper) pph.RoundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(r)
		metrics.CloudProviderCallCount.
			WithLabelValues(
				"ccee",
				r.URL.Path,
				fmt.Sprintf("%d", resp.StatusCode),
				cceemeta.GetCceeRegion(resp.Request.Context()),
				fmt.Sprintf("%s/%s", cceemeta.GetCceeDomain(resp.Request.Context()), cceemeta.GetCceeProject(resp.Request.Context())),
			).
			Inc()
		return resp, err
	}
}

func monitoredHttpClient() *http.Client {
	c := http.DefaultClient
	transport := http.DefaultTransport
	return &http.Client{
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
		Timeout:       c.Timeout,
		Transport:     instrumentCounter(transport),
	}
}
