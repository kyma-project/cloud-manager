package looper

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/feature"
)

func TestMain(m *testing.M) {
	err := feature.Initialize(context.Background(), logr.Discard(),
		feature.WithFile("testdata/ffAllEnabled.yaml"))
	if err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}
