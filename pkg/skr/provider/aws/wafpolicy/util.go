package wafpolicy

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
)

// ScopeRegional returns the REGIONAL scope for WebACLs
func ScopeRegional() wafv2types.Scope {
	return wafv2types.ScopeRegional
}

// convertTags converts Cloud Manager metadata to AWS WAF tags
func convertTags(webAcl *cloudresourcesv1beta1.WafPolicy, state *State) []wafv2types.Tag {
	tags := []wafv2types.Tag{
		{
			Key:   aws.String(common.TagCloudManagerName),
			Value: aws.String(state.Name().String()),
		},
		{
			Key:   aws.String(common.TagScope),
			Value: aws.String(state.Scope().Name),
		},
		{
			Key:   aws.String(common.TagShoot),
			Value: aws.String(state.Scope().Spec.ShootName),
		},
	}
	return tags
}
