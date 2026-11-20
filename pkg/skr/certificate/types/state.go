package types

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
)

// CertificateData holds certificate material extracted from the referenced Secret.
type CertificateData struct {
	Certificate      []byte // tls.crt - leaf PEM, single block
	PrivateKey       []byte // tls.key
	CertificateChain []byte // ca.crt (optional)
}

// State is the shared interface for Certificate state that all provider-specific states extend.
type State interface {
	commonscope.State
	ObjAsCertificate() *cloudresourcesv1beta1.Certificate
	GetData() *CertificateData
	GetSecretHash() string
}
