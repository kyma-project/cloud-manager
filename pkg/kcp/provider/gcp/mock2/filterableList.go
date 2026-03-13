package mock2

import (
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/mock2/filter"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

type FilterableListItem[T any] struct {
	Obj  T
	Name gcputil.NameDetail
}

type FilterableList[T any] struct {
	items        []FilterableListItem[T]
	filterEngine *filter.FilterEngine[T]
}

type FilterableListOption func(o *filterableListOptions)

type filterableListOptions struct {
	createFilter bool
}

func MustNewFilterableList[T any](opts ...FilterableListOption) *FilterableList[T] {
	return util.Must(NewFilterableList[T](opts...))
}

func WithoutFilter(o *filterableListOptions) {
	o.createFilter = false
}

func WithFilter(o *filterableListOptions) {
	o.createFilter = true
}

func NewFilterableList[T any](opts ...FilterableListOption) (*FilterableList[T], error) {
	o := &filterableListOptions{}
	opts = append([]FilterableListOption{
		func(o *filterableListOptions) {
			o.createFilter = true
		},
	}, opts...)
	for _, opt := range opts {
		opt(o)
	}
	var fe *filter.FilterEngine[T]
	var err error
	if o.createFilter {
		fe, err = filter.NewFilterEngine[T]()
		if err != nil {
			return nil, err
		}
	}
	return &FilterableList[T]{
		filterEngine: fe,
	}, nil
}

func MapFilterableList[A any, B any](in *FilterableList[A], objMapping func(A) B, nameMapping func(gcputil.NameDetail) gcputil.NameDetail) (*FilterableList[B], error) {
	if in == nil {
		return nil, fmt.Errorf(":%w: list to map is nil", common.ErrLogical)
	}
	if objMapping == nil {
		return nil, fmt.Errorf("%w: object mapped callback func is nil", common.ErrLogical)
	}
	result, err := NewFilterableList[B](WithFilter)
	if err != nil {
		return nil, err
	}
	if nameMapping == nil {
		nameMapping = func(in gcputil.NameDetail) gcputil.NameDetail {
			return in
		}
	}
	for _, item := range in.items {
		mappedObj := objMapping(item.Obj)
		mappedName := nameMapping(item.Name)
		result.items = append(result.items, FilterableListItem[B]{
			Obj:  mappedObj,
			Name: mappedName,
		})
	}
	return result, nil
}

func (l *FilterableList[T]) Len() int {
	return len(l.items)
}

func (l *FilterableList[T]) Add(obj T, name gcputil.NameDetail) {
	l.items = append(l.items, FilterableListItem[T]{
		Obj:  obj,
		Name: name,
	})
}

func (l *FilterableList[T]) GetItems() []T {
	return pie.Map(l.items, func(i FilterableListItem[T]) T {
		return i.Obj
	})
}

func (l *FilterableList[T]) FindByName(name gcputil.NameDetail) (T, bool) {
	for _, item := range l.items {
		if item.Name.Equal(name) {
			return item.Obj, true
		}
	}
	var zero T
	return zero, false
}

func (l *FilterableList[T]) FilterByParent(parent gcputil.NameDetail) *FilterableList[T] {
	result := &FilterableList[T]{
		filterEngine: l.filterEngine,
	}
	for _, item := range l.items {
		if item.Name.StartsWith(parent) {
			result.items = append(result.items, item)
		}
	}
	return result
}

func (l *FilterableList[T]) FilterByExpression(f *string) (*FilterableList[T], error) {
	if f == nil {
		return l, nil
	}
	result := &FilterableList[T]{
		filterEngine: l.filterEngine,
	}
	for _, item := range l.items {
		match, err := l.filterEngine.Match(*f, item.Obj)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate filter expression: %w", err)
		}
		if match {
			result.items = append(result.items, item)
		}
	}
	return result, nil
}

func (l *FilterableList[T]) FilterByCallback(f func(l FilterableListItem[T]) bool) *FilterableList[T] {
	result := &FilterableList[T]{
		filterEngine: l.filterEngine,
	}
	for _, item := range l.items {
		if f(item) {
			result.items = append(result.items, item)
		}
	}
	return result
}

func (l *FilterableList[T]) FilterNotByCallback(f func(l FilterableListItem[T]) bool) *FilterableList[T] {
	result := &FilterableList[T]{
		filterEngine: l.filterEngine,
	}
	for _, item := range l.items {
		if !f(item) {
			result.items = append(result.items, item)
		}
	}
	return result
}

func (l *FilterableList[T]) ToIterator() gcpclient.Iterator[T] {
	return &iteratorMocked[T]{
		items: pie.Map(l.items, func(i FilterableListItem[T]) T {
			return i.Obj
		}),
	}
}
