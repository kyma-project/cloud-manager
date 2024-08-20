package mock

import (
	cceeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/client"
	cceenfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/nfsinstance/client"
)

type NfsInstanceClient interface {
	cceenfsinstanceclient.Client
}

type Clients interface {
	NfsInstanceClient
}

type Providers interface {
	NfsInstanceProvider() cceeclient.CceeClientProvider[cceenfsinstanceclient.Client]
}

type Server interface {
	Clients

	Providers

	NfsConfig
}
