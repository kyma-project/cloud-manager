package dsl

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
		secret.Data = map[string][]byte{
			"tls.crt": []byte("-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKZ7VfZPJdqLMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV\n-----END CERTIFICATE-----"),
			"tls.key": []byte("-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDExampleKey==\n-----END PRIVATE KEY-----"),
			"ca.crt":  []byte("-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKZ7VfZPJdqLMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV\n-----END CERTIFICATE-----"),
		}
	}

	err := clnt.Create(ctx, secret)
	return err
}
