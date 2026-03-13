package mock2

type subscription struct {
	Store
	server Server
}

func NewSubscription(server Server, projectId string) Subscription {
	return &subscription{
		Store:  newStore(projectId),
		server: server,
	}
}

var _ Subscription = (*subscription)(nil)

func (s *subscription) Delete() {
	s.server.DeleteSubscription(s.Store.ProjectId())
}
