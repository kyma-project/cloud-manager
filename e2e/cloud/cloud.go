package cloud

import (
	"context"
	"fmt"

	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cloud interface {
	CredentialsEnv() map[string]string
	WorkspaceBuilder(alias string) *TFWorkspaceBuilder
}

func Create(ctx context.Context, gardenClient client.Client, config *e2econfig.ConfigType) (Cloud, error) {
	ch := NewCredentialsHandler(gardenClient, config)
	if err := ch.LoadAllOnce(ctx); err != nil {
		return nil, fmt.Errorf("error loading gardener credentials: %w", err)
	}
	return New(ch, config), nil
}

func New(ch CredentialsHandler, config *e2econfig.ConfigType) Cloud {
	return &cloudImpl{
		ch:     ch,
		config: config,
	}
}

type cloudImpl struct {
	ch     CredentialsHandler
	config *e2econfig.ConfigType
}

func (c *cloudImpl) CredentialsEnv() map[string]string {
	return c.ch.Env()
}

func (c *cloudImpl) WorkspaceBuilder(alias string) *TFWorkspaceBuilder {
	return NewTFWorkspaceBuilder(
		alias,
		c.CredentialsEnv(),
		c.config,
	)
}
