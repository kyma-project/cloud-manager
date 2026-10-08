package mock

import (
	"context"

	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	"github.com/kyma-project/cloud-manager/pkg/util"

	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/googleapi"
)

var _ Server = &server{}

func New() Server {
	return &server{
		serviceUsageStore: &serviceUsageStore{},
		vpcPeeringStore:   &vpcPeeringStore{},
		exposedDataStore: &exposedDataStore{
			ipPool: util.Must(iprangeallocate.NewAddressSpace("33.0.0.0/16")),
		},
	}
}

type server struct {
	*serviceUsageStore
	*vpcPeeringStore
	*exposedDataStore
}

func (s *server) SetSuEnableError(err *googleapi.Error) {
	s.suEnableError = err
}

func (s *server) SetSuDisableError(err *googleapi.Error) {
	s.suDisableError = err
}

func (s *server) SetSuOperationError(err *googleapi.Error) {
	s.suOperationError = err
}

func (s *server) SetSuIsEnabledError(err *googleapi.Error) {
	s.suIsEnabledError = err
}

func (s *server) ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient] {
	return func(ctx context.Context, credentialsFile string) (client.ServiceUsageClient, error) {
		return s, nil
	}
}

func (s *server) VpcPeeringProvider() client.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient] {
	return func(_ string) gcpvpcpeeringclient.VpcPeeringClient {
		return s
	}
}

func (s *server) ExposedDataProvider() client.GcpClientProvider[gcpexposeddataclient.Client] {
	return func(_ string) gcpexposeddataclient.Client {
		return s
	}
}
