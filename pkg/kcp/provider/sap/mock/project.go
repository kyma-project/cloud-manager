package mock

import sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"

var _ Project = (*project)(nil)

type project struct {
	*mainStore

	server Server

	domainName  string
	projectName string
	regionName  string
}

func (p *project) Equals(other Project) bool {
	return p.DomainName() == other.DomainName() &&
		p.ProjectName() == other.ProjectName() &&
		p.RegionName() == other.RegionName()
}

func (p *project) ProviderParams() sapclient.ProviderParams {
	return sapclient.ProviderParams{
		DomainName:  p.domainName,
		ProjectName: p.projectName,
		RegionName:  p.regionName,
	}
}

func (p *project) DomainName() string {
	return p.domainName
}

func (p *project) ProjectName() string {
	return p.projectName
}

func (p *project) RegionName() string {
	return p.regionName
}

func (p *project) Delete() {
	p.server.DeleteProject(p)
}
