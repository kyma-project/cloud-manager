package mock

import (
	"context"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/elliotchance/pie/v2"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func newOperationalInsightsStore(subscription string) *operationalInsightsStore {
	return &operationalInsightsStore{
		subscription: subscription,
	}
}

type operationalInsightsStore struct {
	m            sync.Mutex
	subscription string

	logAnalyticsWorkspaces []*armoperationalinsights.Workspace
}

var _ azureclient.OperationalInsightsClient = (*operationalInsightsStore)(nil)

func (s *operationalInsightsStore) GetLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, options *armoperationalinsights.WorkspacesClientGetOptions) (armoperationalinsights.WorkspacesClientGetResponse, error) {
	result := armoperationalinsights.WorkspacesClientGetResponse{}
	if isContextCanceled(ctx) {
		return result, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := azureutil.NewLogAnalyticsWorkspaceResourceId(s.subscription, resourceGroupName, workspaceName).String()

	for _, ws := range s.logAnalyticsWorkspaces {
		if ptr.Deref(ws.ID, "") == id {
			result.Workspace = *ws
			return result, nil
		}
	}

	return result, azuremeta.NewAzureNotFoundError()
}

func (s *operationalInsightsStore) CreateOrUpdateLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, parameters armoperationalinsights.Workspace, options *armoperationalinsights.WorkspacesClientBeginCreateOrUpdateOptions) (azureclient.Poller[armoperationalinsights.WorkspacesClientCreateOrUpdateResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := azureutil.NewLogAnalyticsWorkspaceResourceId(s.subscription, resourceGroupName, workspaceName).String()

	// remove existing - the updat case
	s.logAnalyticsWorkspaces = pie.FilterNot(s.logAnalyticsWorkspaces, func(ws *armoperationalinsights.Workspace) bool {
		return ptr.Deref(ws.ID, "") == id
	})

	ws := util.Must(util.Clone(&parameters))
	ws.ID = new(id)
	ws.Name = new(workspaceName)
	s.logAnalyticsWorkspaces = append(s.logAnalyticsWorkspaces, ws)

	return NewPollerMock(armoperationalinsights.WorkspacesClientCreateOrUpdateResponse{
		Workspace: *util.Must(util.Clone(ws)),
	}, nil, ""), nil
}

func (s *operationalInsightsStore) DeleteLogAnalyticsWorkspace(ctx context.Context, resourceGroupName string, workspaceName string, options *armoperationalinsights.WorkspacesClientBeginDeleteOptions) (azureclient.Poller[armoperationalinsights.WorkspacesClientDeleteResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := azureutil.NewLogAnalyticsWorkspaceResourceId(s.subscription, resourceGroupName, workspaceName).String()

	s.logAnalyticsWorkspaces = pie.FilterNot(s.logAnalyticsWorkspaces, func(ws *armoperationalinsights.Workspace) bool {
		return ptr.Deref(ws.ID, "") == id
	})

	return NewPollerMock(armoperationalinsights.WorkspacesClientDeleteResponse{}, nil, ""), nil
}
