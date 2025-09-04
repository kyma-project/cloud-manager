package mock

import (
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData/client"
	sapnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance/client"
)

type NfsInstanceClient interface {
	sapnfsinstanceclient.Client
}

type Clients interface {
	NfsInstanceClient
}

type Providers interface {
	NfsInstanceProvider() sapclient.SapClientProvider[sapnfsinstanceclient.Client]
	ExposedDataProvider() sapclient.SapClientProvider[sapexposeddataclient.Client]
}

type Server interface {
	Clients

	Providers

	NfsConfig
}
