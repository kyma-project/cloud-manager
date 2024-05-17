package feature

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
)

func DecorateLogger(ctx context.Context, logger logr.Logger) logr.Logger {
	ffCtx := ContextFromCtx(ctx)
	if ffCtx == nil {
		return logger
	}
	return attributesToLogger(ffCtx.GetCustom(), logger)
}

var loggerKeys = []types.Key{
	types.KeyFeature,
	types.KeyPlane,
	types.KeyProvider,
	types.KeyBrokerPlan,
	types.KeyGlobalAccount,
	types.KeySubAccount,
	types.KeyKyma,
	types.KeyShoot,
	types.KeyRegion,
	types.KeyObjKindGroup,
	types.KeyCrdKindGroup,
	types.KeyBusolaKindGroup,
}

func attributesToLogger(attributes map[string]interface{}, logger logr.Logger) logr.Logger {
	var loggerValues []any
	for _, lk := range loggerKeys {
		v, ok := attributes[lk]
		if ok {
			s := fmt.Sprintf("%v", v)
			loggerValues = append(loggerValues, lk, s)
		}
	}
	if len(loggerValues) > 0 {
		newLogger := logger.WithValues(loggerValues...)
		return newLogger
	}
	return logger
}
