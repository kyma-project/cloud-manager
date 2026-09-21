package config

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credentials "github.com/aliyun/credentials-go/credentials"
)

// ApplyCredentials configures an AliCloud openapi client Config for authentication.
//
// When assumeRoleArn is non-empty, it configures a ram_role_arn credential: the
// static accessKeyId/accessKeySecret (the central Cloud Manager technical user)
// only bootstrap an STS AssumeRole into the per-account role, and all API calls
// run under the assumed role's auto-refreshed session credentials. This mirrors
// the AWS client's NewSkrConfig assume-role path.
//
// When assumeRoleArn is empty, it falls back to raw AK/SK auth (used by envtest,
// the in-memory mock, and single-account setups without a role).
//
// The openapi SDK selects raw AK/SK auth whenever both AccessKeyId and
// AccessKeySecret are set (darabonba-openapi client.Init), so the assume-role
// path deliberately leaves those unset and provides Credential instead.
func ApplyCredentials(config *openapi.Config, accessKeyId, accessKeySecret, assumeRoleArn string) error {
	if assumeRoleArn == "" {
		config.AccessKeyId = tea.String(accessKeyId)
		config.AccessKeySecret = tea.String(accessKeySecret)
		return nil
	}

	cred, err := credentials.NewCredential(&credentials.Config{
		Type:            tea.String("ram_role_arn"),
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		RoleArn:         tea.String(assumeRoleArn),
		RoleSessionName: tea.String(AlicloudConfig.RoleSessionName),
	})
	if err != nil {
		return err
	}
	config.Credential = cred
	return nil
}
