package dsl

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
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

// GenerateTestCertificateWithChain generates a certificate where tls.crt contains
// both the leaf certificate and the intermediate CA concatenated (Gardener-style chain),
// with no separate ca.crt key. This mirrors secrets like kyma-gateway-certs.
func GenerateTestCertificateWithChain(commonName, org string) (tlsCrtPEM []byte, keyPEM []byte, err error) {
	certPEM, keyPEM, caPEM, err := GenerateTestCertificate(commonName, org)
	if err != nil {
		return nil, nil, err
	}
	// Concatenate leaf cert + intermediate CA into tls.crt (Gardener convention)
	tlsCrtPEM = append(certPEM, caPEM...)
	return tlsCrtPEM, keyPEM, nil
}
