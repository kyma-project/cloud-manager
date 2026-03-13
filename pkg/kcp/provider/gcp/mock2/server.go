package mock2

import (
	"fmt"
	"sync"

	"github.com/elliotchance/pie/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func New() Server {
	return &server{}
}

type server struct {
	m sync.Mutex

	subscriptions []Subscription
}

// Providers START - add new provider methods below ====================================================

func (s *server) ExposedDataProvider() gcpclient.GcpClientProvider[gcpexposeddataclient.Client] {
	return func(projectId string) gcpexposeddataclient.Client {
		return s.GetSubscription(projectId)
	}
}

// Providers END - add new provider methods above ===========================================

func (s *server) NewSubscription(prefix string) Subscription {
	s.m.Lock()
	defer s.m.Unlock()

	if prefix == "" {
		prefix = "e2e"
	}
	name := fmt.Sprintf("%s-%s", prefix, util.RandomString(10))

	sub := NewSubscription(s, name)
	s.subscriptions = append(s.subscriptions, sub)

	return sub
}

// GetSubscription returns previously created subscription with the given projectId. If there is no subscription with
// such projectId, nil is returned, intentionally so that reconciler fails, which would indicate invalid test setup
// and a signal to developer to create the subscription at the beginning of the test.
func (s *server) GetSubscription(projectId string) Subscription {
	s.m.Lock()
	defer s.m.Unlock()

	for _, p := range s.subscriptions {
		if p.ProjectId() == projectId {
			return p
		}
	}
	return nil
}

func (s *server) DeleteSubscription(projectId string) {
	s.m.Lock()
	defer s.m.Unlock()

	s.subscriptions = pie.FilterNot(s.subscriptions, func(s Subscription) bool {
		return s.ProjectId() == projectId
	})
}

var _ Server = (*server)(nil)
