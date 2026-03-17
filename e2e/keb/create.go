package keb

import (
	"context"
	"fmt"
	"time"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Create(ctx context.Context, config *e2econfig.ConfigType) (Keb, error) {
	kcpClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: commonscheme.KcpScheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kcp client: %w", err)
	}

	if err := e2elib.InitializeKcp(ctx, kcpClient, config); err != nil {
		return nil, fmt.Errorf("failed to initialize kcp: %w", err)
	}

	gardenClientFactory := e2elib.NewGardenClientFactory(kcpClient, config.GardenNamespace)
	gardenClient, err := gardenClientFactory.CreateClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create garden client: %w", err)
	}

	skrManagerFactory := NewSkrManagerFactory(kcpClient, clock.RealClock{}, config.KcpNamespace)
	cpl := e2elib.NewGardenCloudProfileLoader(gardenClient, config)
	skrKubeconfigProvider := e2elib.NewGardenSkrKubeconfigProvider(gardenClient, 10*time.Hour, config.GardenNamespace)

	return NewKeb(kcpClient, gardenClient, skrManagerFactory, cpl, skrKubeconfigProvider, config), nil
}
