package nuke

import (
	"context"
	"strings"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func deleteCertificates(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, rks := range state.ProviderResources {
		if rks.Kind == "AwsCertificate" && rks.Provider == cloudcontrolv1beta1.ProviderAws {
			for _, obj := range rks.Objects {
				resource := obj.(AwsCertificateResource)
				arn := ptr.Deref(resource.Summary.CertificateArn, "")

				err := state.awsClient.DeleteCertificate(ctx, arn)
				if err != nil {
					if awsmeta.IsNotFound(err) {
						logger.Info("Certificate already deleted", "arn", arn)
						continue
					}
					if isResourceInUseException(err) {
						logger.Info("Certificate in use, cannot delete", "arn", arn)
						continue // Expected - ALB/CloudFront may be using it
					}
					logger.Error(err, "Error requesting AWS Certificate deletion", "arn", arn)
					// Continue anyway (non-fatal error pattern)
				} else {
					logger.Info("Requested AWS Certificate deletion", "arn", arn)
				}
			}
		}
	}

	return nil, nil
}

func isResourceInUseException(err error) bool {
	if err == nil {
		return false
	}
	// Check if error contains "ResourceInUseException"
	return strings.Contains(err.Error(), "ResourceInUseException")
}
