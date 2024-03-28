package composed

import "sigs.k8s.io/controller-runtime/pkg/client"

type ObjectList interface {
	client.ObjectList
	GetItemCount() int
	GetItems() []client.Object
}
