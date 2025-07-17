package mock

import (
	"context"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance/client"
)

func New() Server {
	return &server{
		&nfsStore{},
	}
}

type server struct {
	*nfsStore
}

func (s *server) NfsInstanceProvider() sapclient.SapClientProvider[sapnfsinstanceclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapnfsinstanceclient.Client, error) {
		return s, nil
	}
}
