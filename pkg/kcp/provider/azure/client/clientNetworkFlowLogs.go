package client

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"
)

type NetworkFlowLogsClient interface {
	ListNetworkWatchers(ctx context.Context) ([]*armnetwork.Watcher, error)
	CreateNetworkWatcher(ctx context.Context, resourceGroupName, watcherName string, watcher armnetwork.Watcher) (*armnetwork.Watcher, error)

	GetFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, options *armnetwork.FlowLogsClientGetOptions) (armnetwork.FlowLogsClientGetResponse, error)
	CreateOrUpdateFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, parameters armnetwork.FlowLog, options *armnetwork.FlowLogsClientBeginCreateOrUpdateOptions) (Poller[armnetwork.FlowLogsClientCreateOrUpdateResponse], error)
	DeleteFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, options *armnetwork.FlowLogsClientBeginDeleteOptions) (Poller[armnetwork.FlowLogsClientDeleteResponse], error)
}

type networkFlowLogsClient struct {
	svcWatchers *armnetwork.WatchersClient
	svcFlowLogs *armnetwork.FlowLogsClient
}

func NewNetworkFlowLogsClient(
	svcWatchers *armnetwork.WatchersClient,
	svcFlowLogs *armnetwork.FlowLogsClient,
) NetworkFlowLogsClient {
	return &networkFlowLogsClient{
		svcWatchers: svcWatchers,
		svcFlowLogs: svcFlowLogs,
	}
}

// Watchers ===================================================

func (c *networkFlowLogsClient) ListNetworkWatchers(ctx context.Context) ([]*armnetwork.Watcher, error) {
	pager := c.svcWatchers.NewListAllPager(&armnetwork.WatchersClientListAllOptions{})

	var items []*armnetwork.Watcher

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		items = append(items, page.Value...)
	}

	return items, nil
}

func (c *networkFlowLogsClient) CreateNetworkWatcher(ctx context.Context, resourceGroupName, watcherName string, watcher armnetwork.Watcher) (*armnetwork.Watcher, error) {
	location := ptr.Deref(watcher.Location, "")
	if location == "" {
		return nil, fmt.Errorf("watcher location cannot be empty")
	}
	if resourceGroupName == "" {
		resourceGroupName = "NetworkWatcherRG"
	}
	if watcherName == "" {
		watcherName = fmt.Sprintf("NetworkWatcher_%s", location)
	}
	watcher.Name = &watcherName
	resp, err := c.svcWatchers.CreateOrUpdate(ctx, resourceGroupName, watcherName, watcher, &armnetwork.WatchersClientCreateOrUpdateOptions{})
	if err != nil {
		return nil, err
	}
	return &resp.Watcher, nil
}

// FlowLogs ===========================================

func (c *networkFlowLogsClient) GetFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, options *armnetwork.FlowLogsClientGetOptions) (armnetwork.FlowLogsClientGetResponse, error) {
	return c.svcFlowLogs.Get(ctx, resourceGroupName, networkWatcherName, flowLogName, options)
}

func (c *networkFlowLogsClient) CreateOrUpdateFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, parameters armnetwork.FlowLog, options *armnetwork.FlowLogsClientBeginCreateOrUpdateOptions) (Poller[armnetwork.FlowLogsClientCreateOrUpdateResponse], error) {
	return c.svcFlowLogs.BeginCreateOrUpdate(ctx, resourceGroupName, networkWatcherName, flowLogName, parameters, options)
}

func (c *networkFlowLogsClient) DeleteFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, options *armnetwork.FlowLogsClientBeginDeleteOptions) (Poller[armnetwork.FlowLogsClientDeleteResponse], error) {
	return c.svcFlowLogs.BeginDelete(ctx, resourceGroupName, networkWatcherName, flowLogName, options)
}
