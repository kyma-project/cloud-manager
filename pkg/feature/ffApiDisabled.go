package feature

import (
	"context"
)

const apiDisabledFlagName = "apiDisabled"

var ApiDisabled = &ApiDisabledInfo{}

type ApiDisabledInfo struct{}

func (f *ApiDisabledInfo) Value(ctx context.Context) bool {
	v, err := provider.BoolVariation(apiDisabledFlagName, MustContextFromCtx(ctx), false)
	if err != nil {
		return false
	}
	return v
}
