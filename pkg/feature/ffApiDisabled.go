package feature

import (
	"context"
)

const apiDisabledFlagName = "apiDisabled"

var ApiDisabled = &ApiDisabledInfo{}

type ApiDisabledInfo struct{}

func (f *ApiDisabledInfo) Value(ctx context.Context) bool {
	ffCtx := MustContextFromCtx(ctx)
	if ffCtx.GetCustom()[KeyKindGroup] == "cloudresources.cloud-resources.kyma-project.io" {
		return false
	}
	v, err := provider.BoolVariation(apiDisabledFlagName, ffCtx, false)
	if err != nil {
		return false
	}
	return v
}
