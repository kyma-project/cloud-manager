package mock2

// ReactorFunc is a reactor handler that mutates the object when its created and added to the mock. If it returns true
// it means that the reactor handled the object and no other reactors should be executed.
type ReactorFunc[T any] func(T) bool

type Reactor[T any] interface {
	AddReactor(ReactorFunc[T])
	React(T)
}

var _ Reactor[any] = (*reactorImpl[any])(nil)

func newReactor[T any](reactors ...ReactorFunc[T]) Reactor[T] {
	return &reactorImpl[T]{
		reactors: reactors,
	}
}

type reactorImpl[T any] struct {
	reactors []ReactorFunc[T]
}

func (r *reactorImpl[T]) AddReactor(rf ReactorFunc[T]) {
	r.reactors = append(r.reactors, rf)
}

func (r *reactorImpl[T]) React(obj T) {
	for _, rf := range r.reactors {
		if rf(obj) {
			break
		}
	}
}


