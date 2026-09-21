package wafpolicy

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
)

// ScopeRegional returns the REGIONAL scope for WebACLs
func ScopeRegional() wafv2types.Scope {
	return wafv2types.ScopeRegional
}

// convertTags converts Cloud Manager metadata to AWS WAF tags
func convertTags(webAcl *cloudresourcesv1beta1.WafPolicy, scope *cloudcontrolv1beta1.Scope) []wafv2types.Tag {
	tags := []wafv2types.Tag{
		{
			Key:   aws.String("kyma-project.io/managed-by"),
			Value: aws.String("cloud-manager"),
		},
		{
			Key:   aws.String("kyma-project.io/shoot-name"),
			Value: aws.String(scope.Spec.ShootName),
		},
		{
			Key:   aws.String("kyma-project.io/resource-name"),
			Value: aws.String(fmt.Sprintf("%s/%s", webAcl.Namespace, webAcl.Name)),
		},
	}
	return tags
}
