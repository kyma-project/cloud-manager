package awscertificate

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func parseCertificateData(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	cert := state.ObjAsAwsCertificate()

	// Extract certificate data from Secret
	certificate, ok := state.secret.Data["tls.crt"]
	if !ok || len(certificate) == 0 {
		logger.Error(fmt.Errorf("missing tls.crt"), "Secret missing required key tls.crt")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError("Secret missing required key tls.crt")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	privateKey, ok := state.secret.Data["tls.key"]
	if !ok || len(privateKey) == 0 {
		logger.Error(fmt.Errorf("missing tls.key"), "Secret missing required key tls.key")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError("Secret missing required key tls.key")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Certificate chain is optional
	certificateChain := state.secret.Data["ca.crt"]

	// Parse certificate to extract issuer and serial number
	block, _ := pem.Decode(certificate)
	if block == nil {
		logger.Error(fmt.Errorf("failed to decode PEM"), "Failed to decode certificate PEM")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError("Failed to decode certificate PEM")
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		logger.Error(err, "Failed to parse X.509 certificate")
		return composed.NewStatusPatcherComposed(cert).
			MutateStatus(func(c *cloudresourcesv1beta1.AwsCertificate) {
				c.SetStatusProviderError(fmt.Sprintf("Failed to parse certificate: %v", err))
			}).
			OnSuccess(composed.Requeue).
			Run(ctx, state.Cluster().K8sClient())
	}

	// Format serial number for AWS ACM: hex bytes separated by colons
	// AWS expects format: [0-9a-f]{2}(:[0-9a-f]{2}){1,19}
	serialHex := hex.EncodeToString(x509Cert.SerialNumber.Bytes())
	// Insert colons every 2 characters
	var serialFormatted strings.Builder
	for i := 0; i < len(serialHex); i += 2 {
		if i > 0 {
			serialFormatted.WriteString(":")
		}
		if i+2 <= len(serialHex) {
			serialFormatted.WriteString(serialHex[i : i+2])
		} else {
			// Handle odd-length hex string (prepend with 0)
			serialFormatted.WriteString("0")
			serialFormatted.WriteString(serialHex[i:])
		}
	}

	state.certificateData = &CertificateData{
		Certificate:      certificate,
		PrivateKey:       privateKey,
		CertificateChain: certificateChain,
		SerialNumber:     x509Cert.SerialNumber,
		SerialFormatted:  serialFormatted.String(), // AWS ACM format
	}

	logger = logger.WithValues("serial", state.certificateData.SerialFormatted)

	logger.Info("Certificate data parsed successfully")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
