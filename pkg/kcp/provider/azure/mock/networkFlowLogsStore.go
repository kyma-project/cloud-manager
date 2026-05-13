package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/elliotchance/pie/v2"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func newNetworkFlowLogsStore(subscription string) *networkFlowLogsStore {
	return &networkFlowLogsStore{
		subscription: subscription,
	}
}

type networkFlowLogsStore struct {
	m            sync.Mutex
	subscription string

	watchers []*armnetwork.Watcher
	flowLogs []*armnetwork.FlowLog
}

var _ azureclient.NetworkFlowLogsClient = (*networkFlowLogsStore)(nil)

func (s *networkFlowLogsStore) ListNetworkWatchers(ctx context.Context) ([]*armnetwork.Watcher, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	result := make([]*armnetwork.Watcher, len(s.watchers))
	for _, watcher := range s.watchers {
		result = append(result, util.Must(util.Clone(watcher)))
	}

	return result, nil
}

func (s *networkFlowLogsStore) CreateNetworkWatcher(ctx context.Context, resourceGroupName, watcherName string, watcher armnetwork.Watcher) (*armnetwork.Watcher, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if watcherName == "" {
		return nil, fmt.Errorf("watcher name must not be empty")
	}
	location := ptr.Deref(watcher.Location, "")
	if location == "" {
		return nil, fmt.Errorf("watcher location is required")
	}
	for _, watcher := range s.watchers {
		if ptr.Deref(watcher.Location, "") == location {
			return nil, fmt.Errorf("watcher already exists on location %s", location)
		}
	}

	w := util.Must(util.Clone(&watcher))
	w.ID = new(azureutil.NewNetworkWatcherResourceId(s.subscription, resourceGroupName, watcherName).String())
	w.Name = new(watcherName)

	s.watchers = append(s.watchers, w)

	return util.Must(util.Clone(w)), nil
}

func (s *networkFlowLogsStore) GetFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, options *armnetwork.FlowLogsClientGetOptions) (armnetwork.FlowLogsClientGetResponse, error) {
	result := armnetwork.FlowLogsClientGetResponse{}
	if isContextCanceled(ctx) {
		return result, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := azureutil.NewNetworkFlowLogResourceId(s.subscription, resourceGroupName, networkWatcherName, flowLogName).String()

	for _, flowLog := range s.flowLogs {
		if ptr.Deref(flowLog.ID, "") == id {
			fl := util.Must(util.Clone(flowLog))
			result.FlowLog = *fl
			return result, nil
		}
	}

	return result, azuremeta.NewAzureNotFoundError()
}

func (s *networkFlowLogsStore) CreateOrUpdateFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, parameters armnetwork.FlowLog, options *armnetwork.FlowLogsClientBeginCreateOrUpdateOptions) (azureclient.Poller[armnetwork.FlowLogsClientCreateOrUpdateResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := azureutil.NewNetworkFlowLogResourceId(s.subscription, resourceGroupName, networkWatcherName, flowLogName).String()

	// remove existing if any - the update flow
	s.flowLogs = pie.FilterNot(s.flowLogs, func(fl *armnetwork.FlowLog) bool {
		return ptr.Deref(fl.ID, "") == id
	})

	flowLog := util.Must(util.Clone(&parameters))
	flowLog.ID = &id
	flowLog.Name = new(flowLogName)
	s.flowLogs = append(s.flowLogs, flowLog)

	return NewPollerMock(armnetwork.FlowLogsClientCreateOrUpdateResponse{
		FlowLog: *flowLog,
	}, nil, ""), nil
}

func (s *networkFlowLogsStore) DeleteFlowLog(ctx context.Context, resourceGroupName string, networkWatcherName string, flowLogName string, options *armnetwork.FlowLogsClientBeginDeleteOptions) (azureclient.Poller[armnetwork.FlowLogsClientDeleteResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := azureutil.NewNetworkFlowLogResourceId(s.subscription, resourceGroupName, networkWatcherName, flowLogName).String()

	s.flowLogs = pie.FilterNot(s.flowLogs, func(fl *armnetwork.FlowLog) bool {
		return ptr.Deref(fl.ID, "") == id
	})

	return NewPollerMock(armnetwork.FlowLogsClientDeleteResponse{}, nil, ""), nil
}
