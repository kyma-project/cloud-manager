package util

import (
	"fmt"

	alicloudconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/config"
)

// RoleArn returns the AliCloud RAM role ARN in the form
// "acs:ram::<accountId>:role/<roleName>".
func RoleArn(accountId, roleName string) string {
	return fmt.Sprintf("acs:ram::%s:role/%s", accountId, roleName)
}

// RoleArnDefault returns the RAM role ARN for the given account id using the
// role name configured in AlicloudConfig.AssumeRoleName. It returns an empty
// string when accountId is empty, so callers can fall back to raw AK/SK auth.
func RoleArnDefault(accountId string) string {
	if accountId == "" {
		return ""
	}
	return RoleArn(accountId, alicloudconfig.AlicloudConfig.AssumeRoleName)
}
