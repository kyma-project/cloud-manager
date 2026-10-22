package certificate

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func compareCertificateData(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.certificateDetail == nil {
		logger.Info("Certificate not found in ACM, import needed")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	awsCert, awsChain, err := state.awsClient.GetCertificate(ctx, state.certificateArn)
	if err != nil {
		logger.Error(err, "Error getting certificate from ACM")
		return composed.LogErrorAndReturn(err, "Error getting certificate from ACM", composed.StopWithRequeue, ctx)
	}

	secretCertPEM := string(state.certificateData.Certificate)

	if !arePEMsEqual(awsCert, secretCertPEM) {
		logger.Info("Certificate body differs, update needed")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	secretChainPEM := string(state.certificateData.CertificateChain)

	if len(secretChainPEM) == 0 {
		if !arePEMsEqual(awsChain, secretCertPEM) {
			logger.Info("Certificate chain removed from secret, update needed")
			state.certificateNeedsUpdate = true
			return nil, ctx
		}
	} else if !arePEMsEqual(awsChain, secretChainPEM) {
		logger.Info("Certificate chain differs, update needed")
		state.certificateNeedsUpdate = true
		return nil, ctx
	}

	state.certificateNeedsUpdate = false
	return nil, ctx
}

func arePEMsEqual(pem1, pem2 string) bool {
	return normalizePEM(pem1) == normalizePEM(pem2)
}

func normalizePEM(pem string) string {
	return strings.TrimSpace(strings.ReplaceAll(pem, "\r\n", "\n"))
}
