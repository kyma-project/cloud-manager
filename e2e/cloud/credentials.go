package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CredentialsHandler interface {
	LoadAllOnce(ctx context.Context) error
	IsLoaded() bool
	Env() map[string]string
}

func NewCredentialsHandler(gardenClient client.Client, config *e2econfig.ConfigType) CredentialsHandler {
	return &credentialsHandlerImpl{
		gardenClient: gardenClient,
		config:       config,
		data:         make(map[cloudcontrolv1beta1.ProviderType]cloudCredentials),
	}
}

type cloudCredentials interface {
	toEnv(map[string]string)
}

type AwsCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

func (c *AwsCredentials) toEnv(env map[string]string) {
	env["AWS_ACCESS_KEY_ID"] = c.AccessKeyID
	env["AWS_SECRET_ACCESS_KEY"] = c.SecretAccessKey
	env["AWS_DEFAULT_REGION"] = e2elib.DefaultRegions[cloudcontrolv1beta1.ProviderAws]
}

type AzureCredentials struct {
	ClientID       string
	ClientSecret   string
	TenantID       string
	SubscriptionID string
}

func (c *AzureCredentials) toEnv(env map[string]string) {
	env["ARM_CLIENT_ID"] = c.ClientID
	env["ARM_CLIENT_SECRET"] = c.ClientSecret
	env["ARM_TENANT_ID"] = c.TenantID
	env["ARM_SUBSCRIPTION_ID"] = c.SubscriptionID
}

type GcpCredentials struct {
	ServiceAccountKey string
	Project           string
}

func (c *GcpCredentials) toEnv(env map[string]string) {
	env["GOOGLE_CREDENTIALS"] = c.ServiceAccountKey
	env["GOOGLE_PROJECT"] = c.Project
}

type credentialsHandlerImpl struct {
	m sync.Mutex

	gardenClient client.Client
	config       *e2econfig.ConfigType
	isLoaded     bool
	data         map[cloudcontrolv1beta1.ProviderType]cloudCredentials
}

func (h *credentialsHandlerImpl) Env() map[string]string {
	env := map[string]string{}
	for _, creds := range h.data {
		creds.toEnv(env)
	}
	return env
}

func (h *credentialsHandlerImpl) IsLoaded() bool {
	return h.isLoaded
}

func (h *credentialsHandlerImpl) LoadAllOnce(ctx context.Context) error {
	h.m.Lock()
	defer h.m.Unlock()

	if h.isLoaded {
		return nil
	}

	var result error
	if err := h.LoadSubscriptionCredentials(ctx, h.config.Subscriptions.GetDefaultForProvider(cloudcontrolv1beta1.ProviderAws).Name); err != nil {
		result = multierror.Append(result, fmt.Errorf("aws: %w", err))
	}
	if err := h.LoadSubscriptionCredentials(ctx, h.config.Subscriptions.GetDefaultForProvider(cloudcontrolv1beta1.ProviderAzure).Name); err != nil {
		result = multierror.Append(result, fmt.Errorf("azure: %w", err))
	}
	if err := h.LoadSubscriptionCredentials(ctx, h.config.Subscriptions.GetDefaultForProvider(cloudcontrolv1beta1.ProviderGCP).Name); err != nil {
		result = multierror.Append(result, fmt.Errorf("gcp: %w", err))
	}
	h.isLoaded = true

	return result
}

func (h *credentialsHandlerImpl) LoadSubscriptionCredentials(ctx context.Context, subscriptionName string) error {
	out, err := commongardener.LoadGardenerCloudProviderCredentials(ctx, commongardener.LoadGardenerCloudProviderCredentialsInput{
		Client:      h.gardenClient,
		Namespace:   h.config.GardenNamespace,
		BindingName: subscriptionName,
	})
	if err != nil {
		return fmt.Errorf("error loading subscription credentials: %w", err)
	}

	pt, err := cloudcontrolv1beta1.ParseProviderType(out.Provider)
	if err != nil {
		return fmt.Errorf("error parsing provider type: %w", err)
	}

	var creds cloudCredentials

	switch pt {
	case cloudcontrolv1beta1.ProviderAws:
		creds, err = h.awsCredentials(out.CredentialsData)
	case cloudcontrolv1beta1.ProviderAzure:
		creds, err = h.azureCredentials(out.CredentialsData)
	case cloudcontrolv1beta1.ProviderGCP:
		creds, err = h.gcpCredentials(out.CredentialsData)
	default:
		return fmt.Errorf("unsupported provider type: %s", out.Provider)
	}
	if err != nil {
		return fmt.Errorf("error loading %s credentials: %w", out.Provider, err)
	}

	h.data[pt] = creds

	return nil
}

func (h *credentialsHandlerImpl) awsCredentials(credentialsData map[string]string) (*AwsCredentials, error) {
	accessKeyID, ok := credentialsData["accessKeyID"]
	if !ok {
		return nil, fmt.Errorf("missing credential key %s", "accessKeyID")
	}
	secretAccessKey, ok := credentialsData["secretAccessKey"]
	if !ok {
		return nil, fmt.Errorf("missing credential key %s", "secretAccessKey")
	}

	return &AwsCredentials{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
	}, nil
}

func (h *credentialsHandlerImpl) azureCredentials(credentialsData map[string]string) (*AzureCredentials, error) {
	clientId, ok := credentialsData["clientID"]
	if !ok {
		return nil, fmt.Errorf("missing credential key %s", "clientID")
	}
	clientSecret, ok := credentialsData["clientSecret"]
	if !ok {
		return nil, fmt.Errorf("missing credential key %s", "clientSecret")
	}
	subscriptionID, ok := credentialsData["subscriptionID"]
	if !ok {
		return nil, fmt.Errorf("missing credential key %s", "subscriptionID")
	}
	tenantID, ok := credentialsData["tenantID"]
	if !ok {
		return nil, fmt.Errorf("missing credential key %s", "tenantID")
	}

	return &AzureCredentials{
		ClientID:       clientId,
		ClientSecret:   clientSecret,
		TenantID:       tenantID,
		SubscriptionID: subscriptionID,
	}, nil
}

func (h *credentialsHandlerImpl) gcpCredentials(credentialsData map[string]string) (*GcpCredentials, error) {
	txt, ok := credentialsData["serviceaccount.json"]
	if !ok {
		return nil, fmt.Errorf("missing credential key %s", "serviceaccount.json")
	}
	tmp := map[string]string{}
	err := json.Unmarshal([]byte(txt), &tmp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling gcp credential data: %w", err)
	}
	projectId, ok := tmp["project_id"]
	if !ok {
		return nil, fmt.Errorf("missing project_id in gcp credentials")
	}

	return &GcpCredentials{
		ServiceAccountKey: txt,
		Project:           projectId,
	}, nil
}
