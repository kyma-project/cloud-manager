package client

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type providerSuite struct {
	suite.Suite
	ctx context.Context
}

func (ste *providerSuite) SetupTest() {
	ste.ctx = log.IntoContext(context.Background(), logr.Discard())
}

func (ste *providerSuite) TestGetCachedGcpClient() {
	ctx := context.Background()
	saJsonKeyPath := os.Getenv("GCP_SA_JSON_KEY_PATH")
	err := os.Setenv("GCP_CLIENT_RENEW_DURATION", "500ms")
	assert.Nil(ste.T(), err)
	prevClient := &http.Client{}
	renewed := 0
	for i := 0; i < 33; i++ {
		client, err := GetCachedGcpClient(ctx, saJsonKeyPath)
		assert.Nil(ste.T(), err)
		if prevClient != client {
			renewed++
		}
		time.Sleep(100 * time.Millisecond)
		prevClient = client
	}
	assert.Equal(ste.T(), 7, renewed) //First loot iteration also adds to renewed count. So the result is 1 + totalTime/duration i.e. 1 + 33/5 which is 7
}

func TestProvider(t *testing.T) {
	t.SkipNow() // This test relies on the environment variable GCP_SA_JSON_KEY_PATH and also connection to gcp end point so skipping it for now. If needed can be commented out for manual testing
	suite.Run(t, new(providerSuite))
}
