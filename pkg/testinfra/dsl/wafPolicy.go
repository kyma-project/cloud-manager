package dsl

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateWafPolicy(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.WafPolicy, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(obj)

	// Set default data if not already set
	if obj.Spec.Data == "" {
		obj.Spec.Data = `{
			"DefaultAction": {
				"Allow": {}
			},
			"VisibilityConfig": {
				"SampledRequestsEnabled": true,
				"CloudWatchMetricsEnabled": true,
				"MetricName": "` + obj.Name + `-metrics"
			}
		}`
	}

	err := clnt.Create(ctx, obj)
	return err
}
