package mock

import (
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	gcpnfsinstancev1client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1/client"
	gcpnfsinstancev2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	gcpnfsrestoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpsubnetclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/googleapi"
)

type IpRangeClient interface {
	gcpiprangeclient.ComputeClient
	gcpiprangeclient.ServiceNetworkingClient
}

type NfsClient interface {
	gcpnfsinstancev1client.FilestoreClient
}

type Clients interface {
	IpRangeClient
	NfsClient
	gcpsubnetclient.ComputeClient
	gcpsubnetclient.NetworkConnectivityClient
	gcpexposeddataclient.Client
	gcpnfsinstancev2client.FilestoreClient
}

type Providers interface {
	ServiceNetworkingClientProvider() client.ClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	// GcpClientProvider versions for NEW pattern
	ServiceNetworkingClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient]
	ComputeClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ComputeClient]
	ComputeClientProvider() client.ClientProvider[gcpiprangeclient.ComputeClient]
	OldComputeClientProvider() client.ClientProvider[gcpiprangeclient.OldComputeClient] // For v2 legacy
	SubnetComputeClientProvider() client.GcpClientProvider[gcpsubnetclient.ComputeClient]
	SubnetNetworkConnectivityProvider() client.GcpClientProvider[gcpsubnetclient.NetworkConnectivityClient]
	SubnetRegionOperationsClientProvider() client.GcpClientProvider[gcpsubnetclient.RegionOperationsClient]
	FilestoreClientProvider() client.ClientProvider[gcpnfsinstancev1client.FilestoreClient]
	FilestoreClientProviderV2() client.GcpClientProvider[gcpnfsinstancev2client.FilestoreClient]
	ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient]
	FilerestoreClientProvider() client.ClientProvider[gcpnfsrestoreclient.FileRestoreClient]
	FileBackupClientProvider() client.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]
	FileBackupClientProviderV2() client.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
	VpcPeeringProvider() client.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient]
	MemoryStoreProviderFake() client.GcpClientProvider[gcpredisinstanceclient.MemorystoreClient]
	MemoryStoreClusterProviderFake() client.GcpClientProvider[gcpredisclusterclient.MemorystoreClusterClient]
	ExposedDataProvider() client.GcpClientProvider[gcpexposeddataclient.Client]
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

	IpRangeClientUtils

	MemoryStoreClientFakeUtils
	MemoryStoreClusterClientFakeUtils

	RegionalOperationsClientFakeUtils

	VpcPeeringMockClientUtils

	ExposedDataConfig

	FileBackupClientFakeUtils

	FileBackupClientFakeUtilsV2

	FilestoreClientFakeUtils
}
