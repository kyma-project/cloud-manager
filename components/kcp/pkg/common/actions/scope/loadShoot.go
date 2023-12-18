package scope

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadShoot(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	shoot, err := state.GardenerClient.Shoots(state.ShootNamespace).Get(ctx, state.ShootName, metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("error getting shoot: %w", err)
		logger.Error(err, "Error loading shoot")
		return err, nil
	}

	state.Shoot = shoot

	logger.Info("Shoot loaded")

	return nil, nil
}
