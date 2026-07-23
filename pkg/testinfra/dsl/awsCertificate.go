package dsl

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GenerateTestCertificate generates a valid CA-signed certificate chain for testing.
// Returns server certificate PEM, private key PEM, CA certificate PEM, and error.
// The CA certificate is a self-signed root that signs the server certificate.
func GenerateTestCertificate(commonName, org string) (certPEM []byte, keyPEM []byte, caPEM []byte, err error) {
	// Generate CA certificate (root)
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	caSerialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, nil, err
	}

	caTemplate := x509.Certificate{
		SerialNumber: caSerialNumber,
		Subject: pkix.Name{
			Organization: []string{org + " CA"},
			CommonName:   "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10 years for CA
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            1,
	}

	// Self-sign the CA certificate
	caDerBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDerBytes})

	// Parse CA certificate for signing
	caCert, err := x509.ParseCertificate(caDerBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	// Generate server certificate
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	serverSerialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, nil, err
	}

	serverTemplate := x509.Certificate{
		SerialNumber: serverSerialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
			CommonName:   commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Sign server certificate with CA
	serverDerBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDerBytes})

	serverPrivBytes, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		return nil, nil, nil, err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverPrivBytes})

	return certPEM, keyPEM, caPEM, nil
}

func CreateAwsCertificate(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsCertificate, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)
	return err
}

func CreateCertificateSecret(ctx context.Context, clnt client.Client, secret *corev1.Secret, opts ...ObjAction) error {
	NewObjActions(opts...).ApplyOnObject(secret)

	// Set default type if not set
	if secret.Type == "" {
		secret.Type = corev1.SecretTypeTLS
	}

	// Set default certificate data if not provided
	if len(secret.Data) == 0 {
		certPEM, keyPEM, caPEM, err := GenerateTestCertificate("test.example.com", "Test Org")
		if err != nil {
			return err
		}

		secret.Data = map[string][]byte{
			"tls.crt": certPEM,
			"tls.key": keyPEM,
			"ca.crt":  caPEM,
		}
	}

	err := clnt.Create(ctx, secret)
	return err
}
