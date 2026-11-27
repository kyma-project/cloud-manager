package nuke

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	"k8s.io/utils/ptr"
)

func loadCertificates(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	scopeName := state.ObjAsNuke().Spec.Scope.Name

	// List all certificates
	summaries, err := state.awsClient.ListCertificates(ctx)
	if err != nil {
		logger.Error(err, "Error listing AWS Certificates")
		state.hadErrors = true
		return nil, ctx
	}

	// Create resource kind state
	rks := &nuketypes.ProviderResourceKindState{
		Kind:     "Certificate",
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
			state.hadErrors = true
			continue
		}

		// Check if this Certificate belongs to our scope
		if hasCertificateTag(tags, common.TagScope, scopeName) {
			rks.Objects = append(rks.Objects, CertificateResource{Summary: summary})
			logger.Info("Found Certificate for scope", "arn", arn, "scope", scopeName)
		}
	}

	state.ProviderResources = append(state.ProviderResources, rks)

	logger.Info("Loaded AWS Certificates for nuke", "count", len(rks.Objects), "scope", scopeName)
	return nil, ctx
}
