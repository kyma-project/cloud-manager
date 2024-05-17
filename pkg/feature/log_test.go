package feature

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr/funcr"
	"github.com/kyma-project/cloud-manager/pkg/feature/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLog(t *testing.T) {

	var savedLogs []map[string]interface{}

	logger := funcr.NewJSON(func(obj string) {
		data := map[string]interface{}{}
		_ = json.Unmarshal([]byte(obj), &data)
		savedLogs = append(savedLogs, data)
	}, funcr.Options{})

	ctx := context.Background()
	ctx = ContextBuilderFromCtx(ctx).
		Feature(types.FeaturePeering).
		Provider("aws").
		BrokerPlan("trial").
		Build(ctx)

	logger = DecorateLogger(ctx, logger)

	logger.Info("first")
	logger.Info("second")

	assert.Len(t, savedLogs, 2)
	assert.Equal(t, types.FeaturePeering, savedLogs[0][types.KeyFeature])
	assert.Equal(t, "aws", savedLogs[0][types.KeyProvider])
	assert.Equal(t, "trial", savedLogs[0][types.KeyBrokerPlan])
}
