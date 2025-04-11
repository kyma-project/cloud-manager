package client

import (
	"context"
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
		method := "?"
		ctx := context.Background()
		if r.URL != nil {
			method = r.URL.Path
			ctx = r.Context()
		}
		responseCode := "0"
		if resp != nil {
			responseCode = fmt.Sprintf("%d", resp.StatusCode)
			if resp.Request != nil {
				ctx = resp.Request.Context()
			}
		}
		metrics.CloudProviderCallCount.
			WithLabelValues(
				"ccee",
				method,
				responseCode,
				cceemeta.GetCceeRegion(ctx),
				fmt.Sprintf("%s/%s", cceemeta.GetCceeDomain(ctx), cceemeta.GetCceeProject(ctx)),
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
