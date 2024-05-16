package feature

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr/funcr"
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
		Feature(FeaturePeering).
		Provider("aws").
		BrokerPlan("trial").
		Build(ctx)

	logger = DecorateLogger(ctx, logger)

	logger.Info("first")
	logger.Info("second")

	assert.Len(t, savedLogs, 2)
	assert.Equal(t, FeaturePeering, savedLogs[0][KeyFeature])
	assert.Equal(t, "aws", savedLogs[0][KeyProvider])
	assert.Equal(t, "trial", savedLogs[0][KeyBrokerPlan])
}
