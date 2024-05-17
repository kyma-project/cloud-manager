package feature

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"strings"
)

const apiDisabledFlagName = "apiDisabled"

var ApiDisabled = &ApiDisabledInfo{}

type ApiDisabledInfo struct{}

func (f *ApiDisabledInfo) Value(ctx context.Context) bool {
	ffCtx := MustContextFromCtx(ctx)
	if strings.Contains(
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyAllKindGroups]),
		"cloudresources.cloud-resources.kyma-project.io",
	) {
		return false
	}
	v, err := provider.BoolVariation(apiDisabledFlagName, ffCtx, false)
	if err != nil {
		return false
	}
	return v
}
