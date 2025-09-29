package mock

import (
	"context"

	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData/client"
	sapiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange/client"
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

func (s *server) IpRangeProvider() sapclient.SapClientProvider[sapiprangeclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapiprangeclient.Client, error) {
		return s, nil
	}
}

func (s *server) NfsInstanceProvider() sapclient.SapClientProvider[sapnfsinstanceclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapnfsinstanceclient.Client, error) {
		return s, nil
	}
}

func (s *server) ExposedDataProvider() sapclient.SapClientProvider[sapexposeddataclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapexposeddataclient.Client, error) {
		return s, nil
	}
}
