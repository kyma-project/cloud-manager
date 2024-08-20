package mock

import (
	"context"
	cceeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/client"
	cceenfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/nfsinstance/client"
)

func New() Server {
	return &server{
		&nfsStore{},
	}
}

type server struct {
	*nfsStore
}

func (s *server) NfsInstanceProvider() cceeclient.CceeClientProvider[cceenfsinstanceclient.Client] {
	return func(ctx context.Context, pp cceeclient.ProviderParams) (cceenfsinstanceclient.Client, error) {
		return s, nil
	}
}
