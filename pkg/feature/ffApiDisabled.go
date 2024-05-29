package feature

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"strings"
)

const apiDisabledFlagName = "apiDisabled"

var ApiDisabled = &apiDisabledInfo{}

type apiDisabledInfo struct{}

func (f *apiDisabledInfo) Value(ctx context.Context) bool {
	ffCtx := MustContextFromCtx(ctx)
	if strings.Contains(
		util.CastInterfaceToString(ffCtx.GetCustom()[types.KeyAllKindGroups]),
		"cloudresources.cloud-resources.kyma-project.io",
	) {
		return false
	}
	v := provider.BoolVariation(ctx, apiDisabledFlagName, false)
	return v
}
