package client

import (
	"iter"

	"github.com/googleapis/gax-go/v2/iterator"
)

type Iterator[T any] interface {
	Next() (T, error)
}

type IteratorWithAll[T any] interface {
	Iterator[T]
	All() iter.Seq2[T, error]
}

func NewIteratorWithAll[T any](it Iterator[T]) IteratorWithAll[T] {
	return &iteratorWithAllImpl[T]{
		it: it,
	}
}

type iteratorWithAllImpl[T any] struct {
	it Iterator[T]
}

func (i *iteratorWithAllImpl[T]) Next() (T, error) {
	return i.it.Next()
}

func (i *iteratorWithAllImpl[T]) All() iter.Seq2[T, error] {
	return iterator.RangeAdapter(i.it.Next)
}
