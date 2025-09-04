package mock

import (
	"context"
	"sync"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	"google.golang.org/api/googleapi"
)

type computeClientFake struct {
	mutex   sync.Mutex
	subnets map[string]*computepb.Subnetwork

	operationsClientUtils RegionalOperationsClientFakeUtils
}

func (computeClientFake *computeClientFake) CreateSubnet(ctx context.Context, request client.CreateSubnetRequest) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	computeClientFake.mutex.Lock()
	defer computeClientFake.mutex.Unlock()

	name := subnet.GetSubnetFullName(request.ProjectId, request.Region, request.Name)

	computeClientFake.subnets[name] = &computepb.Subnetwork{
		Name:                  &name,
		Region:                &request.Region,
		IpCidrRange:           &request.Cidr,
		Network:               &request.Network,
		PrivateIpGoogleAccess: googleapi.Bool(request.PrivateIpGoogleAccess),
		Purpose:               googleapi.String(request.Purpose),
	}

	opKey := computeClientFake.operationsClientUtils.AddRegionOperation(request.Name)

	return opKey, nil
}

func (computeClientFake *computeClientFake) GetSubnet(ctx context.Context, request client.GetSubnetRequest) (*computepb.Subnetwork, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	computeClientFake.mutex.Lock()
	defer computeClientFake.mutex.Unlock()

	name := subnet.GetSubnetFullName(request.ProjectId, request.Region, request.Name)

	if instance, ok := computeClientFake.subnets[name]; ok {
		return instance, nil
	}

	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}
}

func (computeClientFake *computeClientFake) DeleteSubnet(ctx context.Context, request client.DeleteSubnetRequest) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	computeClientFake.mutex.Lock()
	defer computeClientFake.mutex.Unlock()

	name := subnet.GetSubnetFullName(request.ProjectId, request.Region, request.Name)

	if _, ok := computeClientFake.subnets[name]; ok {
		delete(computeClientFake.subnets, name)
		return nil
	}

	return &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}
}
