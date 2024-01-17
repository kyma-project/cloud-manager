package composed

import "context"

func MarkedForDeletionPredicate(ctx context.Context, state State) bool {
	if state.Obj() == nil {
		// TODO: log a warning
		return false
	}
	if state.Obj().GetDeletionTimestamp().IsZero() {
		return false
	}
	return true
}
