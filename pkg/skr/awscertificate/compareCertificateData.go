package awscertificate

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func compareCertificateData(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// If no ARN, need to import
	if state.ObjAsAwsCertificate().Status.Arn == "" {
		logger.Info("Certificate needs import (no ARN)")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	// If certificate not found in ACM (loadCertificate cleared it), need to reimport
	if state.certificateDetail == nil {
		logger.Info("Certificate not found in ACM, reimport needed")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	// Get certificate data from ACM
	awsCert, awsChain, err := state.awsClient.GetCertificate(ctx, state.ObjAsAwsCertificate().Status.Arn)
	if err != nil {
		logger.Error(err, "Error getting certificate from ACM")
		return composed.LogErrorAndReturn(err, "Error getting certificate from ACM", composed.StopWithRequeue, ctx)
	}

	// Compare certificate bodies (PEM format)
	secretCertPEM := string(state.certificateData.Certificate)

	if !arePEMsEqual(awsCert, secretCertPEM) {
		logger.Info("Certificate body differs, update needed")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	// Compare certificate chains (optional)
	secretChainPEM := string(state.certificateData.CertificateChain)

	if !arePEMsEqual(awsChain, secretChainPEM) {
		logger.Info("Certificate chain differs, update needed")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	// No changes detected
	logger.Info("Certificate matches AWS, no update needed")
	state.certificateNeedsUpdate = false
	return nil, ctx
}

// arePEMsEqual normalizes and compares PEM data
func arePEMsEqual(pem1, pem2 string) bool {
	return normalizePEM(pem1) == normalizePEM(pem2)
}

// normalizePEM removes leading/trailing whitespace and normalizes line endings
func normalizePEM(pem string) string {
	return strings.TrimSpace(strings.ReplaceAll(pem, "\r\n", "\n"))
}
