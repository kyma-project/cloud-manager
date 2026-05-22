package nuke

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func loadCertificates(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Get Cloud Manager Scope (Kyma cluster metadata)
	scopeName := state.Scope().Name

	// List all certificates
	summaries, err := state.awsClient.ListCertificates(ctx)
	if err != nil {
		logger.Error(err, "Error listing AWS Certificates")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	// Create resource kind state
	rks := &nuketypes.ProviderResourceKindState{
		Kind:     "AwsCertificate",
		Provider: cloudcontrolv1beta1.ProviderAws,
		Objects:  []nuketypes.ProviderResourceObject{},
	}

	// Filter and add Certificates belonging to this scope
	for _, summary := range summaries {
		arn := ptr.Deref(summary.CertificateArn, "")
		if arn == "" {
			continue
		}

		// Get tags for this Certificate
		tags, err := state.awsClient.ListCertificateTags(ctx, arn)
		if err != nil {
			logger.Error(err, "Error getting tags for Certificate", "arn", arn)
			continue // Skip this Certificate but continue processing others
		}

		// Check if this Certificate belongs to our scope
		if hasCertificateTag(tags, common.TagScope, scopeName) {
			rks.Objects = append(rks.Objects, AwsCertificateResource{Summary: summary})
			logger.Info("Found Certificate for scope", "arn", arn, "scope", scopeName)
		}
	}

	state.ProviderResources = append(state.ProviderResources, rks)

	logger.Info("Loaded AWS Certificates for nuke", "count", len(rks.Objects), "scope", scopeName)
	return nil, nil
}
