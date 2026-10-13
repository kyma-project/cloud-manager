package certificate

import (
	"context"

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
	Certificate      []byte // tls.crt
	PrivateKey       []byte // tls.key
	CertificateChain []byte // ca.crt (optional)
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
