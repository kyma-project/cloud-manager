package mock

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpiprangev3client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	gcpnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	gcpnfsrestoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/googleapi"
)

type IpRangeClient interface {
	gcpiprangeclient.ComputeClient
	gcpiprangeclient.ServiceNetworkingClient
}

type NfsClient interface {
	gcpnfsinstanceclient.FilestoreClient
}

type Clients interface {
	IpRangeClient
	NfsClient
	gcpiprangev3client.ComputeClient
	gcpiprangev3client.NetworkConnectivityClient
}

type Providers interface {
	ServiceNetworkingClientProvider() client.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	ComputeClientProvider() client.ClientProvider[gcpiprangeclient.ComputeClient]
	ComputeClientProviderV3() client.ClientProvider[gcpiprangev3client.ComputeClient]
	NetworkConnectivityProviderV3() client.ClientProvider[gcpiprangev3client.NetworkConnectivityClient]
	FilestoreClientProvider() client.ClientProvider[gcpnfsinstanceclient.FilestoreClient]
	ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient]
	FilerestoreClientProvider() client.ClientProvider[gcpnfsrestoreclient.FileRestoreClient]
	FileBackupClientProvider() client.ClientProvider[gcpnfsbackupclient.FileBackupClient]
	VpcPeeringProvider() cloudclient.ClientProvider[gcpvpcpeeringclient.VpcPeeringClient]
	MemoryStoreProviderFake() client.ClientProvider[gcpredisinstanceclient.MemorystoreClient]
	MemoryStoreClusterProviderFake() client.ClientProvider[gcpredisclusterclient.MemorystoreClusterClient]
}

// ClientErrors is an interface for setting errors on the mock client to simulate Hyperscaler API errors
type ClientErrors interface {
	SetCreateError(err *googleapi.Error)
	SetPatchError(err *googleapi.Error)
	SetDeleteError(err *googleapi.Error)
	SetGetError(err *googleapi.Error)
	SetOperationError(err *googleapi.Error)
}

type Server interface {
	Clients

	Providers

	ClientErrors

	MemoryStoreClientFakeUtils
	MemoryStoreClusterClientFakeUtils

	VpcPeeringMockClientUtils
}
