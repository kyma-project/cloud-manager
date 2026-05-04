package dsl

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsWebAcl(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsWebAcl, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(obj)

	// Set required fields if not already set
	if obj.Spec.DefaultAction.Allow == nil && obj.Spec.DefaultAction.Block == nil {
		obj.Spec.DefaultAction = cloudresourcesv1beta1.DefaultActionAllow()
	}
	if obj.Spec.VisibilityConfig == nil {
		obj.Spec.VisibilityConfig = &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               obj.Name + "-metrics",
			SampledRequestsEnabled:   true,
		}
	}

	err := clnt.Create(ctx, obj)
	return err
}
