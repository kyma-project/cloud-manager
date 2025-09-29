package mock

import (
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData/client"
	sapiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange/client"
	sapnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance/client"
)

type NfsInstanceClient interface {
	sapnfsinstanceclient.Client
}

type IpRangeClient interface {
	sapiprangeclient.Client
}

type Clients interface {
	NfsInstanceClient
	IpRangeClient
}

type Providers interface {
	IpRangeProvider() sapclient.SapClientProvider[sapiprangeclient.Client]
	NfsInstanceProvider() sapclient.SapClientProvider[sapnfsinstanceclient.Client]
	ExposedDataProvider() sapclient.SapClientProvider[sapexposeddataclient.Client]
}

type Server interface {
	Clients

	Providers

	NfsConfig
}
