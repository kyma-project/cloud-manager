package metrics

import (
	"context"
	"fmt"
	sdkmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	smithymiddleware "github.com/aws/smithy-go/middleware"
	"github.com/aws/smithy-go/transport/http"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"time"
)

type awsRequestMetricTuple struct {
	ServiceName   string
	OperationName string
	Region        string
	Latency       time.Duration
	ResponseCode  int
	Subscription  string
}

func awsReportMetrics(metrics *awsRequestMetricTuple) {
	CloudProviderCallCount.WithLabelValues(
		CloudProviderAWS,
		fmt.Sprintf("%s/%s", metrics.ServiceName, metrics.OperationName),
		fmt.Sprintf("%d", metrics.ResponseCode),
		metrics.Region,
		metrics.Subscription,
	).Inc()
}

func AwsReportMetricsMiddleware() smithymiddleware.DeserializeMiddleware {
	// look at https://github.com/aws/aws-sdk-go-v2/issues/1744
	reportRequestMetrics := smithymiddleware.DeserializeMiddlewareFunc("ReportRequestMetrics", func(
		ctx context.Context, in smithymiddleware.DeserializeInput, next smithymiddleware.DeserializeHandler,
	) (
		out smithymiddleware.DeserializeOutput, metadata smithymiddleware.Metadata, err error,
	) {
		requestMadeTime := time.Now()
		out, metadata, err = next.HandleDeserialize(ctx, in)
		if err != nil {
			return out, metadata, err
		}

		responseStatusCode := -1
		switch resp := out.RawResponse.(type) {
		case *http.Response:
			responseStatusCode = resp.StatusCode
		}

		latency := time.Since(requestMadeTime)
		metrics := awsRequestMetricTuple{
			ServiceName:   sdkmiddleware.GetServiceID(ctx),
			OperationName: sdkmiddleware.GetOperationName(ctx),
			Region:        sdkmiddleware.GetRegion(ctx),
			Latency:       latency,
			ResponseCode:  responseStatusCode,
			Subscription:  awsmeta.GetAwsAccountId(ctx),
		}
		awsReportMetrics(&metrics)

		return out, metadata, nil
	})

	return reportRequestMetrics
}
