package mock

import (
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData/client"
	sapiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange/client"
	sapnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance/client"
	sapvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/vpcnetwork/client"
)

type Clients interface {
	sapclient.NetworkClient
	sapclient.PortClient
	sapclient.RouterClient
	sapclient.ShareClient
	sapclient.SubnetClient
}

type Providers interface {
	IpRangeProvider() sapclient.SapClientProvider[sapiprangeclient.Client]
	NfsInstanceProvider() sapclient.SapClientProvider[sapnfsinstanceclient.Client]
	ExposedDataProvider() sapclient.SapClientProvider[sapexposeddataclient.Client]
	VpcNetworkProvider() sapclient.SapClientProvider[sapvpcnetworkclient.Client]
}

type Config interface {
	NfsConfig
}

type Server interface {
	Providers

	NewProject() Project
	GetProject(domainName, project, region string) Project
}

type Project interface {
	DomainName() string
	ProjectName() string
	RegionName() string
	ProviderParams() sapclient.ProviderParams

	Clients
	Config
}
