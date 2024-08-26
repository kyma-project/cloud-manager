package client

import (
	"context"
	cceeconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/ccee/config"
)

type ProviderParams struct {
	ProjectDomainName string `json:"projectDomainName" yaml:"projectDomainName"`
	ProjectName       string `json:"projectName" yaml:"projectName"`
	RegionName        string `json:"regionName" yaml:"regionName"`

	Username string
	Password string
}

func NewProviderParamsFromConfig(cfg *cceeconfig.CCEEConfigStruct) ProviderParams {
	return ProviderParams{
		Username: cfg.Username,
		Password: cfg.Password,
	}
}

func (pp ProviderParams) WithDomain(v string) ProviderParams {
	pp.ProjectDomainName = v
	return pp
}

func (pp ProviderParams) WithProject(v string) ProviderParams {
	pp.ProjectName = v
	return pp
}

func (pp ProviderParams) WithRegion(v string) ProviderParams {
	pp.RegionName = v
	return pp
}

type CceeClientProvider[T any] func(ctx context.Context, pp ProviderParams) (T, error)
