package mock2

import (
	"context"

	"cloud.google.com/go/securitycentermanagement/apiv1/securitycentermanagementpb"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ gcpclient.SecurityCenterManagementClient = (*store)(nil)

func (s *store) GetSecurityCenterService(_ context.Context, name string) (*securitycentermanagementpb.SecurityCenterService, error) {
	s.m.Lock()
	defer s.m.Unlock()

	state, ok := s.sccServices[name]
	if !ok {
		state = securitycentermanagementpb.SecurityCenterService_INHERITED
	}
	return &securitycentermanagementpb.SecurityCenterService{
		Name:                     name,
		IntendedEnablementState:  state,
		EffectiveEnablementState: state,
	}, nil
}

func (s *store) UpdateSecurityCenterService(_ context.Context, svc *securitycentermanagementpb.SecurityCenterService, _ *fieldmaskpb.FieldMask) (*securitycentermanagementpb.SecurityCenterService, error) {
	s.m.Lock()
	defer s.m.Unlock()

	s.sccServices[svc.Name] = svc.IntendedEnablementState
	return &securitycentermanagementpb.SecurityCenterService{
		Name:                     svc.Name,
		IntendedEnablementState:  svc.IntendedEnablementState,
		EffectiveEnablementState: svc.IntendedEnablementState,
	}, nil
}
