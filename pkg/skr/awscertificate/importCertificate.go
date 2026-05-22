package awscertificate

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func importCertificate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsAwsCertificate()

	// Skip if certificate doesn't need update
	if !state.certificateNeedsUpdate {
		logger.Info("Certificate update not needed, skipping import")
		return nil, ctx
	}

	logger.Info("Importing certificate to ACM")

	// Build ImportCertificateInput
	input := &acm.ImportCertificateInput{
		Certificate: state.certificateData.Certificate,
		PrivateKey:  state.certificateData.PrivateKey,
		Tags:        convertTags(cert, state.Scope()),
	}

	// Add certificate ARN if updating existing certificate
	if cert.Status.Arn != "" {
		input.CertificateArn = ptr.To(cert.Status.Arn)
	}

	// Add certificate chain if provided
	if len(state.certificateData.CertificateChain) > 0 {
		input.CertificateChain = state.certificateData.CertificateChain
	}

	// Import certificate
	arn, err := state.awsClient.ImportCertificate(ctx, input)
	if err != nil {
		logger.Error(err, "Error importing certificate to ACM")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError("Error importing certificate to ACM: " + err.Error())
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Store ARN in state (will be used by updateStatus action)
	state.certificateArn = arn

	logger.Info("Certificate imported to ACM successfully", "arn", arn)
	return nil, ctx
}

func convertTags(cert *cloudresourcesv1beta1.AwsCertificate, scope *cloudcontrolv1beta1.Scope) []acmtypes.Tag {
	tags := []acmtypes.Tag{
		{
			Key:   ptr.To("Name"),
			Value: ptr.To(cert.Name),
		},
		{
			Key:   ptr.To("kyma-project.io/managed-by"),
			Value: ptr.To("cloud-manager"),
		},
		{
			Key:   ptr.To(common.TagScope),
			Value: &scope.Name,
		},
		{
			Key:   ptr.To(common.TagShoot),
			Value: &scope.Spec.ShootName,
		},
	}

	// Add user-defined labels as tags
	for k, v := range cert.Labels {
		tags = append(tags, acmtypes.Tag{
			Key:   ptr.To(k),
			Value: ptr.To(v),
		})
	}

	return tags
}
