package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/httptransport"
	"github.com/go-viper/mapstructure/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

type serviceAccountCredentials struct {
	ProjectId               string `mapstructure:"project_id"`
	PrivateKeyId            string `mapstructure:"private_key_id"`
	PrivateKey              string `mapstructure:"private_key"`
	ClientEmail             string `mapstructure:"client_email"`
	ClientId                string `mapstructure:"client_id"`
	AuthUri                 string `mapstructure:"auth_uri"`
	TokenUri                string `mapstructure:"token_uri"`
	AuthProviderX509CertUrl string `mapstructure:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `mapstructure:"client_x509_cert_url"`
	UniverseDomain          string `mapstructure:"universe_domain"`
}

func loadServiceAccountCredentials(credentialsFile string) (*serviceAccountCredentials, error) {
	content, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	tpy, ok := data["type"]
	if !ok || tpy != "service_account" {
		return nil, errors.New(`expected "service_account" type in credentials file`)
	}

	creds := &serviceAccountCredentials{}
	if err := mapstructure.Decode(data, creds); err != nil {
		return nil, err
	}

	return creds, nil
}

type ReloadingSaKeyTokenProvider struct {
	opts *ReloadingSaKeyTokenProviderOptions
}

type ReloadingSaKeyTokenProviderOptions struct {
	CredentialsFile string
	Scopes          []string
}

func (o *ReloadingSaKeyTokenProviderOptions) Validate() error {
	var result error

	_, err := loadServiceAccountCredentials(o.CredentialsFile)
	if err != nil {
		result = multierror.Append(result, err)
	}

	if len(o.Scopes) == 0 {
		result = multierror.Append(result, errors.New("scopes is empty"))
	}

	return result
}

func NewReloadingSaKeyTokenProvider(opts *ReloadingSaKeyTokenProviderOptions) *ReloadingSaKeyTokenProvider {
	return &ReloadingSaKeyTokenProvider{opts: opts}
}

func (p *ReloadingSaKeyTokenProvider) Token(ctx context.Context) (*auth.Token, error) {
	creds, err := loadServiceAccountCredentials(p.opts.CredentialsFile)
	if err != nil {
		return nil, err
	}

	opts := &auth.Options2LO{
		Email:        creds.ClientEmail,
		PrivateKey:   []byte(creds.PrivateKey),
		TokenURL:     creds.TokenUri,
		PrivateKeyID: creds.PrivateKeyId,
		//Subject:       "",
		Scopes: p.opts.Scopes,
		//Expires: 5 * time.Minute,
		//Audience:      "",
		//PrivateClaims: nil,
		Client: defaultClient(),
		//UseIDToken: false,
		//Logger:     nil,
	}

	tp2lo, err := auth.New2LOTokenProvider(opts)
	if err != nil {
		return nil, err
	}

	t, err := tp2lo.Token(ctx)
	if err != nil {
		return nil, err
	}

	return t, nil
}

type clonableTransport interface {
	Clone() *http.Transport
}

func defaultClient() *http.Client {
	if transport, ok := http.DefaultTransport.(clonableTransport); ok {
		return &http.Client{
			Transport: transport.Clone(),
			Timeout:   30 * time.Second,
		}
	}

	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   30 * time.Second,
	}
}

func NewHttpClient(opts *ReloadingSaKeyTokenProviderOptions) (*http.Client, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	tp := auth.NewCachedTokenProvider(
		NewReloadingSaKeyTokenProvider(opts),
		&auth.CachedTokenProviderOptions{},
	)

	httpClient, err := httptransport.NewClient(&httptransport.Options{
		Credentials: auth.NewCredentials(&auth.CredentialsOptions{
			TokenProvider: tp,
		}),
	})

	return httpClient, err
}

// ======================================

type ReloadingSaKeyTokenProviderOptionsBuilder struct {
	credentialsFile string
	scopes          []string
}

func NewReloadingSaKeyTokenProviderOptionsBuilder(credentialsFile string) *ReloadingSaKeyTokenProviderOptionsBuilder {
	return &ReloadingSaKeyTokenProviderOptionsBuilder{credentialsFile: credentialsFile}
}

func (b *ReloadingSaKeyTokenProviderOptionsBuilder) WithScopes(scopes []string) *ReloadingSaKeyTokenProviderOptionsBuilder {
	b.scopes = scopes
	return b
}

func (b *ReloadingSaKeyTokenProviderOptionsBuilder) BuildOptions() *ReloadingSaKeyTokenProviderOptions {
	return &ReloadingSaKeyTokenProviderOptions{
		CredentialsFile: b.credentialsFile,
		Scopes:          b.scopes,
	}
}

func (b *ReloadingSaKeyTokenProviderOptionsBuilder) BuildHttpClient() (*http.Client, error) {
	opts := b.BuildOptions()
	return NewHttpClient(opts)
}

func (b *ReloadingSaKeyTokenProviderOptionsBuilder) MustBuildHttpClient() *http.Client {
	return util.Must(b.BuildHttpClient())
}

func (b *ReloadingSaKeyTokenProviderOptionsBuilder) BuildTokenProvider() auth.TokenProvider {
	return auth.NewCachedTokenProvider(
		NewReloadingSaKeyTokenProvider(b.BuildOptions()),
		&auth.CachedTokenProviderOptions{},
	)
}
