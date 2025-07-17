package client

import (
	"context"
	"fmt"
	"net/http"

	sapmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/meta"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
	pph "github.com/prometheus/client_golang/prometheus/promhttp"
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
				"openstack",
				method,
				responseCode,
				sapmeta.GetSapRegion(ctx),
				fmt.Sprintf("%s/%s", sapmeta.GetSapDomain(ctx), sapmeta.GetSapProject(ctx)),
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
