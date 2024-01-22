package registry

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ForInput struct {
	object client.Object
}

type ForOption interface {
	// ApplyToFor applies this configuration to the given for input.
	ApplyToFor(*ForInput)
}

type WatchesInput struct {
	object       client.Object
	eventHandler handler.EventHandler
	predicates   []predicate.Predicate
}

type WatchesOption interface {
	// ApplyToWatches applies this configuration to the given watches options.
	ApplyToWatches(*WatchesInput)
}
