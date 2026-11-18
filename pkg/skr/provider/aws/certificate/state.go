package certificate

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	certificatetypes "github.com/kyma-project/cloud-manager/pkg/skr/certificate/types"
	certificateclient "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/certificate/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

type CertificateData struct {
	Certificate      []byte // tls.crt - leaf PEM, single block, ready for ACM import
	PrivateKey       []byte // tls.key
	CertificateChain []byte // ca.crt (optional)
	X509             *x509.Certificate
}

type State struct {
	certificatetypes.State
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client]
	env               abstractions.Environment

	awsClient              certificateclient.Client
	roleName               string
	secret                 *corev1.Secret
	certificateDetail      *acmtypes.CertificateDetail
	certificateData        *CertificateData
	certificateNeedsUpdate bool
	certificateArn         string
}

type StateFactory interface {
	NewState(ctx context.Context, certState certificatetypes.State) (*State, error)
}

func NewStateFactory(
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client],
	env abstractions.Environment,
) StateFactory {
	return &stateFactory{
		awsClientProvider: awsClientProvider,
		env:               env,
	}
}

type stateFactory struct {
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client]
	env               abstractions.Environment
}

func (f *stateFactory) NewState(_ context.Context, certState certificatetypes.State) (*State, error) {
	return &State{
		State:             certState,
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}, nil
}

// formatSerial formats a certificate serial number as colon-separated lowercase hex,
// matching the format ACM uses in CertificateDetail.Serial.
func formatSerial(serialBytes []byte) string {
	h := hex.EncodeToString(serialBytes)
	var sb strings.Builder
	for i := 0; i < len(h); i += 2 {
		if i > 0 {
			sb.WriteByte(':')
		}
		sb.WriteString(h[i : i+2])
	}
	return sb.String()
}

func (s *State) convertTags() []acmtypes.Tag {
	tags := []acmtypes.Tag{
		{
			Key:   aws.String("Name"),
			Value: aws.String(s.ObjAsCertificate().Status.Id),
		},
		{
			Key:   aws.String(common.TagCloudManagerName),
			Value: aws.String(s.Name().String()),
		},
		{
			Key:   aws.String(common.TagScope),
			Value: aws.String(s.Scope().Name),
		},
		{
			Key:   ptr.To(common.TagShoot),
			Value: aws.String(s.Scope().Spec.ShootName),
		},
	}

	return tags
}
