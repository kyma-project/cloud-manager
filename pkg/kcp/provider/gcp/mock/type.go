package mock

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	cloudclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	backupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	nfsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	restoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	memoryStoreClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/googleapi"
)

type IpRangeClient interface {
	iprangeclient.ComputeClient
	iprangeclient.ServiceNetworkingClient
}

type NfsClient interface {
	nfsclient.FilestoreClient
}

type Clients interface {
	IpRangeClient
	NfsClient
}

type Providers interface {
	ServiceNetworkingClientProvider() client.ClientProvider[iprangeclient.ServiceNetworkingClient]
	ComputeClientProvider() client.ClientProvider[iprangeclient.ComputeClient]
	FilestoreClientProvider() client.ClientProvider[nfsclient.FilestoreClient]
	ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient]
	FilerestoreClientProvider() client.ClientProvider[restoreclient.FileRestoreClient]
	FileBackupClientProvider() client.ClientProvider[backupclient.FileBackupClient]
	VpcPeeringSkrProvider() cloudclient.ClientProvider[gcpvpcpeeringclient.VpcPeeringClient]
	MemoryStoreProviderFake() client.ClientProvider[memoryStoreClient.MemorystoreClient]
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
}
