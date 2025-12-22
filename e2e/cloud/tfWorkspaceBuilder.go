package cloud

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

type TFWorkspaceBuilder struct {
	ws *tfWorkspace
}

func NewTFWorkspaceBuilder(alias string, env map[string]string, config *e2econfig.ConfigType) *TFWorkspaceBuilder {
	b := &TFWorkspaceBuilder{
		ws: &tfWorkspace{
			alias:   alias,
			rootDir: config.TfWorkspaceDir,
			tfCmd:   config.TfCmd,
			name:    util.RandomString(8),
			env:     env,
		},
	}
	b.ws.data.Module.Variables = map[string]string{}
	return b
}

func (b *TFWorkspaceBuilder) WithSource(source string) *TFWorkspaceBuilder {
	parts := strings.SplitN(source, " ", 2)
	b.ws.data.Module.Source = parts[0]
	if len(parts) == 2 {
		b.ws.data.Module.Version = parts[1]
	}
	return b
}

func (b *TFWorkspaceBuilder) WithProvider(provider string) *TFWorkspaceBuilder {
	parts := strings.SplitN(provider, " ", 2)
	name := filepath.Base(parts[0])
	p := TfTemplateProvider{
		Name:   name,
		Source: parts[0],
	}
	if len(parts) == 2 {
		p.Version = parts[1]
	}
	b.ws.data.Providers = append(b.ws.data.Providers, p)
	return b
}

func (b *TFWorkspaceBuilder) WithVariable(k, v string) *TFWorkspaceBuilder {
	b.ws.data.Module.Variables[k] = v
	return b
}

func (b *TFWorkspaceBuilder) Validate() error {
	var result error
	if b.ws.data.Module.Source == "" {
		result = multierror.Append(result, errors.New("missing module source"))
	}
	for x, p := range b.ws.data.Providers {
		if p.Name == "" {
			result = multierror.Append(result, fmt.Errorf("missing provider %d name", x))
		}
		if p.Source == "" {
			result = multierror.Append(result, fmt.Errorf("missing provider %d source", x))
		}
		if p.Version == "" {
			result = multierror.Append(result, fmt.Errorf("missing provider %d version", x))
		}
	}
	return result
}

func (b *TFWorkspaceBuilder) Build() TFWorkspace {
	return b.ws
}
