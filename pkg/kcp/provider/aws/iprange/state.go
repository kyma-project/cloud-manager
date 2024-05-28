package iprange

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	v1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/v1"
	v2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/v2"
)

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (composed.State, error)
}

func NewStateFactory(skrProvider awsclient.SkrClientProvider[iprangeclient.Client]) StateFactory {
	return &generealStateFactory{
		v1StateFactory: v1.NewStateFactory(skrProvider),
		v2StateFactory: v2.NewStateFactory(skrProvider),
	}
}

type generealStateFactory struct {
	v1StateFactory v1.StateFactory
	v2StateFactory v2.StateFactory
}

func (f *generealStateFactory) NewState(ctx context.Context, ipRangeState iprangetypes.State, logger logr.Logger) (composed.State, error) {
	return nil, errors.New("logical error - not implemented")
	//return f.v1StateFactory.NewState(ctx, ipRangeState, logger)
}
