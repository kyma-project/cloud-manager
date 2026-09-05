package awscertificate

import (
	"context"
	"math/big"

	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	certificateclient "github.com/kyma-project/cloud-manager/pkg/skr/awscertificate/client"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonscope.State
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client]
	env               abstractions.Environment

	awsClient              certificateclient.Client
	roleName               string
	secret                 *corev1.Secret              // Loaded Secret
	certificateDetail      *acmtypes.CertificateDetail // Loaded from ACM
	certificateData        *CertificateData            // Parsed from Secret
	certificateNeedsUpdate bool                        // True if certificate data differs
	certificateArn         string                      // ARN of imported certificate
}

type CertificateData struct {
	Certificate      []byte   // tls.crt
	PrivateKey       []byte   // tls.key
	CertificateChain []byte   // ca.crt (optional)
	SerialNumber     *big.Int // Serial number (raw)
	SerialFormatted  string   // Serial number formatted for AWS ACM (hex with colons)
}

func newStateFactory(
	composedStateFactory composed.StateFactory,
	commonScopeStateFactory commonscope.StateFactory,
	awsClientProvider awsclient.SkrClientProvider[certificateclient.Client],
	env abstractions.Environment,
) *stateFactory {
	return &stateFactory{
		composedStateFactory:    composedStateFactory,
		commonScopeStateFactory: commonScopeStateFactory,
		awsClientProvider:       awsClientProvider,
		env:                     env,
	}
}

type stateFactory struct {
	composedStateFactory    composed.StateFactory
	commonScopeStateFactory commonscope.StateFactory
	awsClientProvider       awsclient.SkrClientProvider[certificateclient.Client]
	env                     abstractions.Environment
}

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
	scopeState, err := f.commonScopeStateFactory.NewState(ctx, req.NamespacedName,
		f.composedStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsCertificate{}),
	)
	if err != nil {
		return nil, err
	}

	return &State{
		State:             scopeState,
		awsClientProvider: f.awsClientProvider,
		env:               f.env,
	}, nil
}

func (s *State) ObjAsAwsCertificate() *cloudresourcesv1beta1.AwsCertificate {
	return s.Obj().(*cloudresourcesv1beta1.AwsCertificate)
}
