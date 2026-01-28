package mock

import (
	"context"
	"fmt"
	"sync"

	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData/client"
	sapiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/iprange/client"
	sapnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/nfsinstance/client"
	sapvpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/vpcnetwork/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func New() Server {
	return &server{}
}

type server struct {
	m sync.Mutex

	projects []*project
}

func (s *server) IpRangeProvider() sapclient.SapClientProvider[sapiprangeclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapiprangeclient.Client, error) {
		p := s.GetProjectByProviderParams(pp)
		if p == nil {
			return nil, fmt.Errorf("no project found for %s", pp.String())
		}
		return s.GetProjectByProviderParams(pp), nil
	}
}

func (s *server) NfsInstanceProvider() sapclient.SapClientProvider[sapnfsinstanceclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapnfsinstanceclient.Client, error) {
		p := s.GetProjectByProviderParams(pp)
		if p == nil {
			return nil, fmt.Errorf("no project found for %s", pp.String())
		}
		return s.GetProjectByProviderParams(pp), nil
	}
}

func (s *server) ExposedDataProvider() sapclient.SapClientProvider[sapexposeddataclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapexposeddataclient.Client, error) {
		p := s.GetProjectByProviderParams(pp)
		if p == nil {
			return nil, fmt.Errorf("no project found for %s", pp.String())
		}
		return s.GetProjectByProviderParams(pp), nil
	}
}

func (s *server) VpcNetworkProvider() sapclient.SapClientProvider[sapvpcnetworkclient.Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapvpcnetworkclient.Client, error) {
		p := s.GetProjectByProviderParams(pp)
		if p == nil {
			return nil, fmt.Errorf("no project found for %s", pp.String())
		}
		return s.GetProjectByProviderParams(pp), nil
	}
}

func (s *server) GetProjectByProviderParams(pp sapclient.ProviderParams) Project {
	return s.GetProject(pp.DomainName, pp.ProjectName, pp.RegionName)
}

func (s *server) GetProject(domainName, project, region string) Project {
	s.m.Lock()
	defer s.m.Unlock()

	for _, p := range s.projects {
		if p.domainName == domainName && p.projectName == project && p.regionName == region {
			return p
		}
	}
	return nil
}

func (s *server) NewProject() Project {
	s.m.Lock()
	defer s.m.Unlock()

	p := &project{
		mainStore:   newMainStore(),
		domainName:  fmt.Sprintf("domain-%s", util.RandomString(8)),
		projectName: fmt.Sprintf("project-%s", util.RandomString(8)),
		regionName:  "eu-de-1",
	}
	s.projects = append(s.projects, p)
	return p
}
