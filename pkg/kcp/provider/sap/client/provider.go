package client

import (
	"context"
	"fmt"

	sapconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/config"
)

type ProviderParams struct {
	DomainName  string `json:"domainName" yaml:"domainName"`
	ProjectName string `json:"projectName" yaml:"projectName"`
	RegionName  string `json:"regionName" yaml:"regionName"`

	Username string
	Password string
}

func NewProviderParamsFromConfig(cfg *sapconfig.SapConfigStruct) ProviderParams {
	return ProviderParams{
		Username: cfg.Username,
		Password: cfg.Password,
	}
}

func (pp ProviderParams) WithDomain(v string) ProviderParams {
	pp.DomainName = v
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

func (pp ProviderParams) String() string {
	return fmt.Sprintf("%s/%s/%s", pp.DomainName, pp.ProjectName, pp.RegionName)
}

type SapClientProvider[T any] func(ctx context.Context, pp ProviderParams) (T, error)
