package mock2

import (
	"fmt"
	"iter"

	"github.com/googleapis/gax-go/v2/iterator"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	otherit "google.golang.org/api/iterator"
)

func IteratorToSlice[T any](it iter.Seq2[T, error]) ([]T, error) {
	var result []T
	for obj, err := range it {
		if err != nil {
			return result, err
		}
		result = append(result, obj)
	}
	return result, nil
}

type iteratorMocked[T any] struct {
	items []T
	err   error
}

func (it *iteratorMocked[T]) Next() (T, error) {
	var zero T
	if it.err != nil {
		return zero, it.err
	}
	if len(it.items) == 0 {
		return zero, otherit.Done
	}
	item := it.items[0]
	it.items = it.items[1:]
	cpy, err := util.JsonClone(item)
	if err != nil {
		return zero, fmt.Errorf("failed to clone item: %w", err)
	}
	return cpy, nil
}

func (it *iteratorMocked[T]) All() iter.Seq2[T, error] {
	return iterator.RangeAdapter(it.Next)
}

var _ gcpclient.Iterator[any] = (*iteratorMocked[any])(nil)
