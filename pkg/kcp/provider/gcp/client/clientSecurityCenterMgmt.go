package client

import (
	"context"

	securitycentermanagement "cloud.google.com/go/securitycentermanagement/apiv1"
	"cloud.google.com/go/securitycentermanagement/apiv1/securitycentermanagementpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type SecurityCenterManagementClient interface {
	GetSecurityCenterService(ctx context.Context, name string) (*securitycentermanagementpb.SecurityCenterService, error)
	UpdateSecurityCenterService(
		ctx context.Context,
		svc *securitycentermanagementpb.SecurityCenterService,
		mask *fieldmaskpb.FieldMask,
	) (*securitycentermanagementpb.SecurityCenterService, error)
}

var _ SecurityCenterManagementClient = (*sccMgmtClient)(nil)

type sccMgmtClient struct {
	inner *securitycentermanagement.Client
}

func (c *sccMgmtClient) GetSecurityCenterService(ctx context.Context, name string) (*securitycentermanagementpb.SecurityCenterService, error) {
	return c.inner.GetSecurityCenterService(ctx, &securitycentermanagementpb.GetSecurityCenterServiceRequest{
		Name: name,
	})
}

func (c *sccMgmtClient) UpdateSecurityCenterService(
	ctx context.Context,
	svc *securitycentermanagementpb.SecurityCenterService,
	mask *fieldmaskpb.FieldMask,
) (*securitycentermanagementpb.SecurityCenterService, error) {
	return c.inner.UpdateSecurityCenterService(ctx, &securitycentermanagementpb.UpdateSecurityCenterServiceRequest{
		SecurityCenterService: svc,
		UpdateMask:            mask,
	})
}
