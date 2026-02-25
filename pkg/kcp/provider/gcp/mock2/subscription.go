package mock2

type subscription struct {
	Store
	projectId string
	server    Server
}

func NewSubscription(server Server, projectId string) Subscription {
	return &subscription{
		Store:     newStore(),
		projectId: projectId,
		server:    server,
	}
}

var _ Subscription = (*subscription)(nil)

func (p *subscription) ProjectId() string {
	return p.projectId
}

func (s *subscription) Delete() {
	s.server.DeleteSubscription(s.projectId)
}
