package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

// NewAlicloudStsGardenClientProvider returns a provider that constructs an
// AlicloudStsClient authenticated with the given (Gardener binding) AK/SK.
// Calling GetCallerIdentity with these credentials returns the account's id,
// stored on the Subscription and later used to build the assume-role ARN.
func NewAlicloudStsGardenClientProvider() awsclient.GardenClientProvider[AlicloudStsClient] {
	return func(ctx context.Context, region, key, secret string) (AlicloudStsClient, error) {
		config := &openapi.Config{
			AccessKeyId:     new(key),
			AccessKeySecret: new(secret),
			RegionId:        new(region),
		}
		// Use the central STS endpoint. GetCallerIdentity is account-global, so the
		// endpoint region is irrelevant to the result; the central domain avoids
		// baking a region assumption into subscription reconciliation.
		config.Endpoint = new("sts.aliyuncs.com")
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
		Query: openapiutil.Query(map[string]any{}),
	}
	params := &openapi.Params{
		Action:      new("GetCallerIdentity"),
		Version:     new("2015-04-01"),
		Protocol:    new("HTTPS"),
		Pathname:    new("/"),
		Method:      new("POST"),
		AuthType:    new("AK"),
		Style:       new("RPC"),
		ReqBodyType: new("formData"),
		BodyType:    new("json"),
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
func accountIdFromCallApiBody(result map[string]any) (string, bool) {
	if inner, ok := result["body"].(map[string]any); ok {
		if id, ok := inner["AccountId"].(string); ok {
			return id, true
		}
	}
	if id, ok := result["AccountId"].(string); ok {
		return id, true
	}
	return "", false
}
