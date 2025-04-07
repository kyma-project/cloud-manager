package mock

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/compute/apiv1/computepb"
	v3iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/v3/client"
	"google.golang.org/api/googleapi"
)

type computeClientFake struct {
	mutex   sync.Mutex
	subnets map[string]*computepb.Subnetwork
}

func (computeClientFake *computeClientFake) CreatePrivateSubnet(ctx context.Context, request v3iprangeclient.CreateSubnetRequest) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	computeClientFake.mutex.Lock()
	defer computeClientFake.mutex.Unlock()

	name := fmt.Sprintf("%s/%s/%s", request.ProjectId, request.Region, request.Name)

	subnet := &computepb.Subnetwork{
		Name:                  &name,
		Region:                &request.Region,
		IpCidrRange:           &request.Cidr,
		Network:               &request.Network,
		PrivateIpGoogleAccess: googleapi.Bool(true),
		Purpose:               googleapi.String("PRIVATE"),
	}

	computeClientFake.subnets[name] = subnet

	return nil
}

func (computeClientFake *computeClientFake) GetSubnet(ctx context.Context, request v3iprangeclient.GetSubnetRequest) (*computepb.Subnetwork, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	computeClientFake.mutex.Lock()
	defer computeClientFake.mutex.Unlock()

	name := fmt.Sprintf("%s/%s/%s", request.ProjectId, request.Region, request.Name)

	if instance, ok := computeClientFake.subnets[name]; ok {
		return instance, nil
	}

	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}
}

func (computeClientFake *computeClientFake) DeleteSubnet(ctx context.Context, request v3iprangeclient.DeleteSubnetRequest) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	computeClientFake.mutex.Lock()
	defer computeClientFake.mutex.Unlock()

	name := fmt.Sprintf("%s/%s/%s", request.ProjectId, request.Region, request.Name)

	if _, ok := computeClientFake.subnets[name]; ok {
		delete(computeClientFake.subnets, name)
		return nil
	}

	return &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}
}
