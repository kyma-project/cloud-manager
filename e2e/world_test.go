package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elliotchance/pie/v2"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"
	"github.com/kyma-project/cloud-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestWorld(t *testing.T) {
	world := NewWorld()
	ctx := context.Background()

	cfg := config.NewConfig(abstractions.NewOSEnvironment())
	InitConfig(cfg)
	cfg.Read()

	err := world.Init(ctx)
	assert.NoError(t, err)

	garden, err := world.ClusterProvider.Garden(ctx)
	assert.NoError(t, err)

	shootList := &gardenertypes.ShootList{}
	err = garden.GetAPIReader().List(ctx, shootList, client.InNamespace(Config.GardenNamespace))
	assert.NoError(t, err)

	for _, shoot := range shootList.Items {
		fmt.Printf("Found shoot: %s\n", shoot.Name)
	}
}

func foo(ctx context.Context, t *testing.T, world *World) {
	kcp, err := world.ClusterProvider.KCP(ctx)
	assert.NoError(t, err)

	err = kcp.AddResources(ctx, []*ResourceDeclaration{
		{
			Alias:      "cm",
			Kind:       "ConfigMap",
			ApiVersion: "v1",
			Name:       "test",
			Namespace:  "default",
		},
	})
	assert.NoError(t, err)

	for {
		data, err := world.EvaluationContext(ctx)
		assert.NoError(t, err)
		fmt.Printf("evaluation context: %v\n", pie.Keys(data))
		//cm := &corev1.ConfigMap{}
		//err = kcp.GetClient().Get(ctx, client.ObjectKey{
		//	Namespace: "default",
		//	Name:      "test",
		//}, cm)
		//
		//if err != nil {
		//	fmt.Printf("error %T: %s\n", err, err)
		//} else {
		//	fmt.Printf("loaded %v\n", cm.Labels)
		//}

		time.Sleep(1 * time.Second)
	}
}
