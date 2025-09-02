package initShared

import (
	"flag"
	"os"
	"path"

	"github.com/kyma-project/cloud-manager/e2e"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	var filename string

	flag.StringVar(&filename, "filename", filename, "Shared runtimes filename, defaults to ${PROJECTROOT}/.runtimes.yaml")
	flag.StringVar(&filename, "f", filename, "Shared runtimes filename, defaults to ${PROJECTROOT}/.runtimes.yaml")

	flag.Parse()

	opts := zap.Options{}
	opts.Development = true
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	cfg := e2e.LoadConfig()

	if filename == "" {
		filename = path.Join(cfg.GetBaseDir(), ".runtimes.yaml")
	}
	sharedState, err := e2e.LoadSharedState(filename)
	if err != nil {
		logger.Error(err, "Failed to load shared state")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()
	ctx = e2e.NewScenarioSession(ctx)

	f := e2e.NewWorldFactory()
	w, err := f.Create(ctx)
	if err != nil {
		logger.Error(err, "Failed to create world")
		os.Exit(1)
	}

	for _, runtimeId := range sharedState.Runtimes {
		logger = logger.WithValues("runtimeID", runtimeId)
		logger.Info("Importing runtime...")
		skr, err := w.SKR().ImportShared(ctx, runtimeId)
		if err != nil {
			logger.Error(err, "Failed to import runtime")
			os.Exit(1)
		}

		logger.WithValues(
			"shoot", skr.ShootName,
			"provider", skr.Provider,
			"alias", skr.Alias,
		).Info("Shared runtime imported")


	}
}
