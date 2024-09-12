package feature

import "context"

const ffAzurePeeringUseNetworkFlagName = "azurePeeringUseNetwork"

var FFAzurePeeringUseNetwork = &ffAzurePeeringUseNetwork{}

type ffAzurePeeringUseNetwork struct{}

func (f *ffAzurePeeringUseNetwork) Value(ctx context.Context) bool {
	v := provider.BoolVariation(ctx, ffAzurePeeringUseNetworkFlagName, true)
	return v
}
