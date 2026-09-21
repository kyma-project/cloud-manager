package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

// NewAlicloudStsGardenClientProvider returns a provider that constructs an
// AlicloudStsClient authenticated with the given (Gardener binding) AK/SK.
// Calling GetCallerIdentity with these credentials returns the account's id,
// stored on the Subscription and later used to build the assume-role ARN.
func NewAlicloudStsGardenClientProvider() awsclient.GardenClientProvider[AlicloudStsClient] {
	return func(ctx context.Context, region, key, secret string) (AlicloudStsClient, error) {
		config := &openapi.Config{
			AccessKeyId:     tea.String(key),
			AccessKeySecret: tea.String(secret),
			RegionId:        tea.String(region),
		}
		config.Endpoint = tea.String(fmt.Sprintf("sts.%s.aliyuncs.com", region))
		c, err := openapi.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("error creating alicloud sts client: %w", err)
		}
		return &alicloudStsClient{c: c}, nil
	}
}

// AlicloudStsClient exposes the single STS operation the subscription reconciler needs.
type AlicloudStsClient interface {
	// GetCallerIdentity returns the AliCloud account id of the credentials in use.
	GetCallerIdentity(ctx context.Context) (string, error)
}

type alicloudStsClient struct {
	c *openapi.Client
}

// GetCallerIdentity calls the STS GetCallerIdentity RPC (API version 2015-04-01)
// via the generic openapi client, avoiding a dedicated STS SDK dependency.
func (c *alicloudStsClient) GetCallerIdentity(ctx context.Context) (string, error) {
	req := &openapi.OpenApiRequest{
		Query: openapiutil.Query(map[string]interface{}{}),
	}
	params := &openapi.Params{
		Action:      tea.String("GetCallerIdentity"),
		Version:     tea.String("2015-04-01"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("RPC"),
		ReqBodyType: tea.String("formData"),
		BodyType:    tea.String("json"),
	}

	body, err := c.c.CallApi(params, req, &util.RuntimeOptions{})
	if err != nil {
		return "", fmt.Errorf("error calling alicloud sts GetCallerIdentity: %w", err)
	}

	// CallApi wraps a json response as {"body": <fields>, "headers": ..., "statusCode": ...},
	// so the STS AccountId field lives under the nested "body" map, not at the top level.
	accountId, ok := accountIdFromCallApiBody(body)
	if !ok || accountId == "" {
		return "", fmt.Errorf("alicloud sts GetCallerIdentity returned no AccountId")
	}
	return accountId, nil
}

// accountIdFromCallApiBody extracts AccountId from a darabonba-openapi CallApi
// result. The json response is nested under the "body" key; we also tolerate a
// flat shape defensively.
func accountIdFromCallApiBody(result map[string]interface{}) (string, bool) {
	if inner, ok := result["body"].(map[string]interface{}); ok {
		if id, ok := inner["AccountId"].(string); ok {
			return id, true
		}
	}
	if id, ok := result["AccountId"].(string); ok {
		return id, true
	}
	return "", false
}
