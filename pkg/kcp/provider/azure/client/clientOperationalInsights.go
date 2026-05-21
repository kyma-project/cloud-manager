package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
)

type OperationalInsightsClient interface {
	GetLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, options *armoperationalinsights.WorkspacesClientGetOptions) (armoperationalinsights.WorkspacesClientGetResponse, error)
	CreateOrUpdateLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, parameters armoperationalinsights.Workspace, options *armoperationalinsights.WorkspacesClientBeginCreateOrUpdateOptions) (Poller[armoperationalinsights.WorkspacesClientCreateOrUpdateResponse], error)
	DeleteLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, options *armoperationalinsights.WorkspacesClientBeginDeleteOptions) (Poller[armoperationalinsights.WorkspacesClientDeleteResponse], error)
}

func NewOperationalInsightsClient(svcWorkspaces *armoperationalinsights.WorkspacesClient) OperationalInsightsClient {
	return &operationalInsightsClient{svcWorkspaces: svcWorkspaces}
}

type operationalInsightsClient struct {
	svcWorkspaces *armoperationalinsights.WorkspacesClient
}

func (c *operationalInsightsClient) GetLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, options *armoperationalinsights.WorkspacesClientGetOptions) (armoperationalinsights.WorkspacesClientGetResponse, error) {
	return c.svcWorkspaces.Get(ctx, resourceGroupName, workspaceName, options)
}

func (c *operationalInsightsClient) CreateOrUpdateLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, parameters armoperationalinsights.Workspace, options *armoperationalinsights.WorkspacesClientBeginCreateOrUpdateOptions) (Poller[armoperationalinsights.WorkspacesClientCreateOrUpdateResponse], error) {
	return c.svcWorkspaces.BeginCreateOrUpdate(ctx, resourceGroupName, workspaceName, parameters, options)
}

func (c *operationalInsightsClient) DeleteLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, options *armoperationalinsights.WorkspacesClientBeginDeleteOptions) (Poller[armoperationalinsights.WorkspacesClientDeleteResponse], error) {
	return c.svcWorkspaces.BeginDelete(ctx, resourceGroupName, workspaceName, options)
}
