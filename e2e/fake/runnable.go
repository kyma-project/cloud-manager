package fake

import "context"

type Runnable struct {
	StartCallCount int
	ErrPre         error
	ErrPost        error
}

func (r *Runnable) Start(ctx context.Context) error {
	r.StartCallCount++
	if r.ErrPre != nil {
		return r.ErrPre
	}
	<-ctx.Done()
	return r.ErrPost
}
